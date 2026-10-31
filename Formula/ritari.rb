class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/cndreisbach/ritari"
  url "https://github.com/cndreisbach/ritari/archive/refs/tags/v4.0.0.tar.gz"
  sha256 "6e5fc37d05349e59639370fb0f3a0bdf5a43e1bb7ce3e933d50b94448ab56799"
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
