class Speed < Formula
  desc "CLI tool to measure internet speed"
  homepage "https://github.com/sibiraj-s/speed"
  head "https://github.com/sibiraj-s/speed.git", branch: "master"
  
  depends_on "go" => :build

  def install
    ENV["CGO_ENABLED"] = "0"
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end
end
