class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/cndreisbach/ritari"
  url "https://github.com/cndreisbach/ritari/archive/refs/tags/v4.1.0.tar.gz"
  sha256 "e711cf351ffb55e2a767f29e0df7b679656d5276c4cd549c5c6f9e14c4c288df"
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
