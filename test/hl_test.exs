defmodule HlTest do
  use ExUnit.Case
  doctest Hl

  test "greets the world" do
    assert Hl.hello() == :world
  end
end
