#!/bin/bash

version=$(cat version.txt)
macosamd64sha=$(cat dist/checksums.txt | grep spot_darwin_amd64 | awk '{print $1}')
macosarm64sha=$(cat dist/checksums.txt | grep spot_darwin_arm64| awk '{print $1}')
linuxamd64sha=$(cat dist/checksums.txt | grep spot_linux_amd64 | awk '{print $1}')
linuxarm64sha=$(cat dist/checksums.txt | grep spot_linux_arm64 | awk '{print $1}')

cat > spotvm.rb <<EOF
class Spotvm < Formula
    desc "spot vm tool"
    homepage "https://github.com/ysicing/spot"
    version "${version}"

    if OS.mac?
      if Hardware::CPU.arm?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_darwin_arm64"
        sha256 "${macosarm64sha}"
      else
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_darwin_amd64"
        sha256 "${macosamd64sha}"
      end
    elsif OS.linux?
      if Hardware::CPU.intel?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_linux_amd64"
        sha256 "${linuxamd64sha}"
      end
      if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_linux_arm64"
        sha256 "${linuxarm64sha}"
      end
    end

    def install
      if OS.mac?
        if Hardware::CPU.intel?
          bin.install "spot_darwin_amd64" => "spotvm"
        else
          bin.install "spot_darwin_arm64" => "spotvm"
        end
      elsif OS.linux?
        if Hardware::CPU.intel?
          bin.install "spot_linux_amd64" => "spotvm"
        else
          bin.install "spot_linux_arm64" => "spotvm"
        end
      end
    end

    test do
      assert_match "spotvm vervion v#{version}", shell_output("#{bin}/spotvm -v")
    end
end

EOF
