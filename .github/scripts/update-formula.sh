#!/usr/bin/env bash
# Prints the Homebrew formula for a release, built from the tagged source.
# Usage: update-formula.sh <version-without-v> <sha256-of-source-tarball>
set -euo pipefail

version=$1
sha256=$2
repo=${GITHUB_REPOSITORY:-cndreisbach/ritari}

cat <<RUBY
class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/${repo}"
  url "https://github.com/${repo}/archive/refs/tags/v${version}.tar.gz"
  sha256 "${sha256}"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/cndreisbach/ritari/internal/cli.version=#{version}"
    system "go", "build", *std_go_args(ldflags:, output: bin/"rit"), "./cmd/rit"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/rit -o text version")
  end
end
RUBY
