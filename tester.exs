defmodule Tester do
    def split(string), do: do_split(to_charlist(string), true, [], [])
  defp do_split([], split, [], res), do: Enum.reverse(res)
  defp do_split([], split, curr, res) do
    res = [Enum.reverse(curr) |> to_string() | res]
    Enum.reverse(res)
  end
  defp do_split([34|tail], true, [], res) do
    do_split(tail, false, [34], res)
  end
  defp do_split([34|tail], true, curr, res) do
    do_split(tail, false, [34], [Enum.reverse(curr) |> to_string() | res])
  end
  defp do_split([34|tail], false, curr, res) do    
    do_split(tail, true, [], [Enum.reverse([34 | curr]) |> to_string() | res])
  end
  defp do_split([32|tail], true, curr, res) do
    do_split(tail, true, [], [Enum.reverse(curr) |> to_string() | res])
  end
  defp do_split([head|tail], true, curr, res) do
    do_split(tail, true, [head | curr], res)
  end
  defp do_split([head|tail], false, curr, res) do
    do_split(tail, false, [head | curr], res)
  end
end