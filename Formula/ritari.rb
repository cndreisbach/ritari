class Ritari < Formula
  desc "Long-term memory for AI agents, kept as OKF markdown files"
  homepage "https://github.com/cndreisbach/ritari"
  url "https://github.com/cndreisbach/ritari/archive/refs/tags/v4.2.0.tar.gz"
  sha256 "d4ce73d551a2146212ed588e8a2b6691db903a0792f5d7c8ba0bd6db885d375a"
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
