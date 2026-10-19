class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/cndreisbach/ritari"
  url "https://github.com/cndreisbach/ritari/archive/refs/tags/v1.3.0.tar.gz"
  sha256 "46426819c6b1ec659e2cf2852b718d5116b0e9e9a1ab9a3af7c3a823f5a468cb"
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
