class Assumer < Formula
  desc "AWS assume role credential wrapper"
  homepage "https://github.com/masahide/assumer"
  url "https://github.com/masahide/assumer/releases/download/v0.1.4/assumer_Darwin_x86_64.tar.gz"
  version "0.1.4"
  sha256 "288b9be6fb24331b43b8f3b28cccbbffbcc6a58d39f10d32a20f2692bcb10358"

  def install
    bin.install "assumer"
  end

  test do
    system "#{bin}/assumer -v"
  end
end
