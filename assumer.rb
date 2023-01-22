# Documentation: https://docs.brew.sh/Formula-Cookbook
#                https://rubydoc.brew.sh/Formula
# PLEASE REMOVE ALL GENERATED COMMENTS BEFORE SUBMITTING YOUR PULL REQUEST!
class Assumer < Formula
  desc "AWS assume role credential wrapper"
  homepage "https://github.com/masahide/assumer"
  license "MIT"
  version "1.0.0"
  on_macos do
    on_arm do
      url "https://github.com/masahide/assumer/releases/download/2.2.2/darwin-arm64.tar.gz"
      sha256 "eac5588b942898da4688ab5f93e5a0c9b97c777b98a9067689dadd91bcc263e3"
    end
    on_intel do
      url "https://github.com/masahide/assumer/releases/download/2.2.2/darwin-amd64.tar.gz"
      sha256 "f6da0af345c60c5d999c51a77eb6733d26a0ade385cd1b4540f5ca35267de1fd"
    end
  end
  on_linux do
    on_arm do
      url "https://github.com/masahide/assumer/releases/download/2.2.2/linux-arm64.tar.gz"
      sha256 "3f00a3938f5c2bc19b56604302566b9799f8fa27242994ae7d7cec17315e05f6"
    end
    on_intel do
      url "https://github.com/masahide/assumer/releases/download/2.2.2/linux-amd64.tar.gz"
	  sha256 "addc8f17d24e53fc1e46039f6fefb25997b04f8b789492617e4ba843a714c6b4"
    end
  end

  # depends_on "cmake" => :build

  def install
    system "chmod", "755", "assumer"
    bin.install "assumer"
  end

  test do
    system "#{bin}/assumer -v"
  end
end
