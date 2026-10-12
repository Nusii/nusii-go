//go:build ignore

// Release helper (see RELEASING.md).
//
//	go run scripts/release.go patch|minor|major   release the next version
//	go run scripts/release.go 1.2.0-beta.1        release an explicit version
//	go run scripts/release.go version             print the current version
//	go run scripts/release.go notes [x.y.z]       print the CHANGELOG.md notes of a version
//
// Releasing refuses to run with uncommitted changes or without notes under
// "## Unreleased" in CHANGELOG.md. It runs gofmt, go vet and go test, writes
// the version to version.go, moves the notes under "## x.y.z (date)", commits
// "Release vx.y.z" and creates the annotated tag vx.y.z. Push it with
// `git push --follow-tags`.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const unreleased = "## Unreleased"

var (
	versionPattern = regexp.MustCompile(`const Version = "([^"]+)"`)
	semverPattern  = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z.-]+)?$`)
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: go run scripts/release.go <patch|minor|major|x.y.z|version|notes [x.y.z]>")
	}
	current := currentVersion()

	switch command := os.Args[1]; command {
	case "version":
		fmt.Println(current)
	case "notes":
		version := current
		if len(os.Args) > 2 {
			version = strings.TrimPrefix(os.Args[2], "v")
		}
		notes := sectionBody(readFile("CHANGELOG.md"), "## "+version)
		if notes == "" {
			fail("CHANGELOG.md has no notes for " + version)
		}
		fmt.Println(notes)
	default:
		release(current, nextVersion(current, command))
	}
}

func release(current, next string) {
	if output(true, "git", "status", "--porcelain") != "" {
		fail("Commit or stash your changes first: the release script needs a clean working tree.")
	}
	changelog := readFile("CHANGELOG.md")
	if sectionBody(changelog, unreleased) == "" {
		fail(`CHANGELOG.md has no notes under "` + unreleased + `". Describe the changes there first.`)
	}
	tag := "v" + next
	if output(false, "git", "tag", "--list", tag) != "" {
		fail("The tag " + tag + " already exists.")
	}

	fmt.Printf("Releasing %s (was %s)\n\n", tag, current)
	if unformatted := output(true, "gofmt", "-l", "."); unformatted != "" {
		fail("These files need gofmt -w:\n" + unformatted)
	}
	run("go", "vet", "./...")
	run("go", "test", "-race", "-count=1", "./...")

	version := readFile("version.go")
	writeFile("version.go", versionPattern.ReplaceAllString(version, `const Version = "`+next+`"`))
	date := time.Now().Format("2006-01-02")
	writeFile("CHANGELOG.md", strings.Replace(changelog, unreleased, unreleased+"\n\n## "+next+" ("+date+")", 1))

	run("git", "add", "version.go", "CHANGELOG.md")
	run("git", "commit", "--quiet", "-m", "Release "+tag)
	run("git", "tag", "-a", tag, "-m", tag)
	fmt.Printf("\nCreated %s. Publish it with:\n\n  git push --follow-tags\n", tag)
}

func nextVersion(current, command string) string {
	parts := semverPattern.FindStringSubmatch(current)
	if parts == nil {
		fail("version.go has an invalid version: " + current)
	}
	major, _ := strconv.Atoi(parts[1])
	minor, _ := strconv.Atoi(parts[2])
	patch, _ := strconv.Atoi(parts[3])
	prerelease := parts[4] != ""

	switch command {
	case "major":
		return fmt.Sprintf("%d.0.0", major+1)
	case "minor":
		return fmt.Sprintf("%d.%d.0", major, minor+1)
	case "patch":
		if prerelease {
			// 1.2.0-beta.1 -> 1.2.0
			return fmt.Sprintf("%d.%d.%d", major, minor, patch)
		}
		return fmt.Sprintf("%d.%d.%d", major, minor, patch+1)
	}

	version := strings.TrimPrefix(command, "v")
	if !semverPattern.MatchString(version) {
		fail(`Unknown command or version "` + command + `". Use patch, minor, major or a version such as 1.2.0 or 1.2.0-beta.1.`)
	}
	return version
}

func currentVersion() string {
	match := versionPattern.FindStringSubmatch(readFile("version.go"))
	if match == nil {
		fail("version.go has no Version constant")
	}
	return match[1]
}

// sectionBody returns the text below a "## " heading, up to the next one.
func sectionBody(changelog, heading string) string {
	lines := strings.Split(changelog, "\n")
	start := -1
	for i, line := range lines {
		if line == heading || strings.HasPrefix(line, heading+" ") {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start+1:end], "\n"))
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fail(fmt.Sprintf("%s %s failed: %v", name, strings.Join(args, " "), err))
	}
}

func output(mustSucceed bool, name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil && mustSucceed {
		fail(fmt.Sprintf("%s %s failed: %v", name, strings.Join(args, " "), err))
	}
	return strings.TrimSpace(string(out))
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err.Error())
	}
	return string(data)
}

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
