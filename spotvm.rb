class Spotvm < Formula
    desc "spot vm tool"
    homepage "https://github.com/ysicing/spot"
    version "0.0.4"

    if OS.mac?
      if Hardware::CPU.arm?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_darwin_arm64"
        sha256 "11798ceb66caf3732ba13cc1f3f44e00e93e4c16a874c829430fe38792e6aa16"
      else
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_darwin_amd64"
        sha256 "b642af83fe7b8fdae4544934bdeed1d9329799b11eb5bf6a2c7c4f15dfa65a26"
      end
    elsif OS.linux?
      if Hardware::CPU.intel?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_linux_amd64"
        sha256 "6ede7302782d9a611fa591b18c994cd8b28b81ac1b2adcf075075e15833611a8"
      end
      if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
        url "https://github.com/ysicing/spot/releases/download/v#{version}/spot_linux_arm64"
        sha256 "f97672bf552a7a8aec9dca0c2cb25a1fb9d09a5c06c2360b5f6699f58e1ce7bc"
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

