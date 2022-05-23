# Proyecto Integrador
# Jacobo Soffer Levy
# A01028653
# 22/05/22
defmodule Hl do
  @moduledoc """
  Functions used to syntax-highlight a Go source file
  """

  @doc """
  Returns a regex that matches Go's keywords
  """
  def keywords do 
    str = [
    "break",
    "case",
    "chan",
    "const",
    "continue",
    "default",
    "defer",
    "else",
    "fallthrough",
    "for",
    "func",
    "go",
    "goto",
    "if",
    "import",
    "interface",
    "map",
    "package",
    "range",
    "return",
    "select",
    "struct",
    "switch",
    "type",
    "var"
  ] |> Enum.join("|")
      ~S"(?<!\w|\d|\-|_|!|=)(" <> str <> ~S")(?!\w|\d|\-|_|!|=)" |> Regex.compile!()
    end

  @doc """
  Returns a regex that matches Go's operators
  """
  def operators do
    str = [
   ~S"\+",
   "-",
   ~S"\*",
   "/",
   "%",
   "&",
   ~S"\|",
   ~S"\^",
   "<<",
   ~S">>",
   ~S"&\^",
   ~S"\+=",
   "-=",
   ~S"\*=",
   "/=",
   "%=",
   "&=",
   ~S"\|=",
   ~S"\^=",
   "<<=",
   ~S">>=",
   ~S"&\^=",
   "&&",
   ~S"\|\|",
   "<-",
   ~S"\+\+",
   "--",
   "==",
   ~S">",
   "<",
   "=",
   "!",
   "~",
   "!=",
   "<=",
   ~S">=",
   ~S"\:=",
   ~S"\.\.\.",
   ~S"\(",
   ~S"\[",
   ~S"\{",
   ",",
   ~S"\.",
   ~S"\)",
   ~S"\]",
   ~S"\}",
   ";",
   ~S"\:"
 ] |> Enum.join("|")
 "(#{str})" |> Regex.compile!()
    end

 # String regex
 def strings(), do: ~r/"[\w\d\s\*\._\$\^\|\/#@?¿ñ'!()=+\-%&[\]{}\\]*"/

 # Numbers regex
 def numbers(), do: ~r/(?<![a-df-gA-DF-G])\d(?![a-df-gA-DF-G])/

 # Variables regex
 def vars(), do: ~r/\s*[A-Za-z_]\w*\s*/


  @doc """
  Returns a regex that matches Go's types
  """
  def types do 
    str = [
    "nil", # not a type but will be treated as such
    "string",
    "bool",
    "int",
    "int8",
    "uint8",
    "byte",
    "int16",
    "uint16",
    "int32",
    "uint32",
    "rune",
    "int64",
    "uint64",
    "uint",
    "uintptr",
    "float32",
    "float64",
    "complex32",
    "complex64"
  ] |> Enum.join("|")
  ~S"(?<!\w|\d|\-|_|!|=)(" <> str <> ~S")(?!\w|\d|\-|_|!|=)" |> Regex.compile!()
    end
  
  @doc """
  Takes a list of strings and returns a formatted HTML document

  ## Examples

      iex> Hl.run_regex(["var", "a", "int"])
      "<span class="kw">var</span><span class="vr">a</span><span class="tp">int</span>"

  """
  def run_regex(strings), do: do_run_regex(strings, [])
  defp do_run_regex([], res), do: Enum.reverse(res) |> Enum.join("")
  defp do_run_regex([head|tail], res) do
    replaced = 
      Regex.replace(operators(), head, "<span class=\"op\">\\0</span>")
    cond do
      Regex.match?(strings(), head) -> 
        do_run_regex(tail, ["<span class=\"st\">#{head}</span>" | res])
      Regex.match?(types(), replaced) ->
          do_run_regex(tail, ["<span class=\"tp\">#{replaced}</span>" | res])
      Regex.match?(keywords(), replaced) ->
          do_run_regex(tail, ["<span class=\"kw\">#{replaced}</span>" | res])
      Regex.match?(vars(), replaced) ->
        do_run_regex(tail, ["<span class=\"vr\">#{replaced}</span>" | res])
      Regex.match?(numbers(), replaced) ->
        do_run_regex(tail, ["<span class=\"nm\">#{replaced}</span>" | res])
      true -> do_run_regex(tail, ["<span>#{replaced}</span>" | res])
    end
  end

  @doc """
  Takes a string and splits it every space, newline or operator.
  It does not split anything that is between quotes.

  ## Examples
  
      iex> Hl.split("Hello \"split elixir\" world")
      ["Hello ", "\"split elixir\"", " ", "world"]
  """
  def split(string), do: do_split(to_charlist(string), true, [], [])
  defp do_split([], _split, [], res), do: Enum.reverse(res)
  defp do_split([], _split, curr, res) do
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
    do_split(tail, true, [], [Enum.reverse([32|curr]) |> to_string() | res])
  end
  defp do_split([10|tail], true, curr, res) do
    do_split(tail, true, [], [Enum.reverse([10|curr]) |> to_string() | res])
  end
  defp do_split([head|tail], true, curr, res) do
    if Regex.match?(operators(), to_string(curr)) do
      do_split(tail, true, [head], [Enum.reverse(curr) |> to_string() | res])
    else
      do_split(tail, true, [head | curr], res)
    end
  end

  defp do_split([head|tail], false, curr, res) do
    do_split(tail, false, [head | curr], res)
  end

  @doc """
  Takes a path to Go source file and output HTML file
  Writes the Go source with syntax highlighting to the
  output file
  """  
  def run(in_filename, out_filename) do
    code = in_filename
    |> File.read!()
    |> split()
    |> run_regex()
    pre = File.read!("pre.html")
    |> String.replace(~S"#{NaiveDateTime.utc_now}", "#{NaiveDateTime.utc_now}")
    post = File.read!("post.html")
    File.write(out_filename, "#{pre}#{code}#{post}")
  end
end
