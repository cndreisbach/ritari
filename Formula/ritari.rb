class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/cndreisbach/ritari"
  url "https://github.com/cndreisbach/ritari/archive/refs/tags/v4.2.1.tar.gz"
  sha256 "ba30ddb4e57bbc2e4edcd3d86699bde15a06463991df4b323a5f03358ff413b2"
  license "MIT"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/cndreisbach/ritari/internal/cli.version=#{version}"
    system "go", "build", *std_go_args(ldflags:, output: bin/"rit"), "./cmd/rit"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/rit -o text version")
  end
end
