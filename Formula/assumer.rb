class Assumer < Formula
  desc "AWS assume role credential wrapper"
  homepage "https://github.com/masahide/assumer"
  url "https://github.com/masahide/assumer/releases/download/v0.1.6/assumer_Darwin_x86_64.tar.gz"
  version "0.1.6"
  sha256 "f8a96c6d287926fa02930ab149a4262b25b1656218367e7bffae76bfa1dcc461"

  def install
    bin.install "assumer"
  end

  test do
    system "#{bin}/assumer -v"
  end
end
