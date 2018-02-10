class Assumer < Formula
  desc "AWS assume role credential wrapper"
  homepage "https://github.com/masahide/assumer"
  url "https://github.com/masahide/assumer/releases/download/v0.1.5/assumer_Darwin_x86_64.tar.gz"
  version "0.1.5"
  sha256 "a057adfe5656ec109fff584b0324939127ae4cdedfb1ad1d8b70ad41e53a82e6"

  def install
    bin.install "assumer"
  end

  test do
    system "#{bin}/assumer -v"
  end
end
