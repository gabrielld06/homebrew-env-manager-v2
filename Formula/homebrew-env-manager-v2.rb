# typed: true
# frozen_string_literal: true

# This file was generated by Homebrew Releaser. DO NOT EDIT.
class HomebrewEnvManagerV2 < Formula
  desc "Environment variable manager for aws amplify, dgo apps and oci object storage,"
  homepage "https://github.com/stanyzra/homebrew-env-manager-v2"
  url "https://github.com/stanyzra/homebrew-env-manager-v2/archive/refs/tags/v0.5.0.tar.gz"
  sha256 "07bf14cbc4aedba0372b3727424f4e7726941b3531583980e1c07b439ef25968"
  license "Apache-2.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", bin/"env-manager-v2"
  end

  test do
    assert_match "0.5.0", shell_output("#{bin}/env-manager-v2 --version")
  end
end
