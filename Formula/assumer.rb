class Assumer < Formula
  desc "AWS assume role credential wrapper"
  homepage "https://github.com/masahide/assumer"
  url "https://github.com/masahide/assumer/releases/download/v0.1.3/assumer_Darwin_x86_64.tar.gz"
  version "0.1.3"
  sha256 "151c14052c78b2bb86a2986c51b4b17f328e2057506c23905153dd2fe311cd21"

  def install
    bin.install "assumer"
  end

  test do
    system "#{bin}/assumer -v"
  end
end
