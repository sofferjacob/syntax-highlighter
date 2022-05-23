def run_trim(string), do: do_run_trim(string, [])

  defp do_run_trim("", res), do: Enum.reverse(res) |> Enum.join("")
  defp do_run_trim(string, res) do
    IO.inspect(res, label: "res")
    cond do
      m = Regex.run(strings(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"st\">#{elem(m, 0)}</span>" | res])
      m = Regex.run(operators(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"op\">#{elem(m, 0)}</span>" | res])
      m = Regex.run(numbers(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"nm\">#{elem(m, 0)}</span>" | res])
      m = Regex.run(types(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"tp\">#{elem(m, 0)}</span>" | res])
      m = Regex.run(keywords(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"kw\">#{elem(m, 0)}</span>" | res])
      m = Regex.run(vars(), string) != nil ->
        do_run_trim(String.trim_leading(string, elem(m, 0)), ["<span class=\"vr\">#{elem(m, 0)}</span>" | res])
      true ->
        do_run_trim("", ["<span>#{string}</span>" | res])
      end
    end
    def run_alt(in_filename, out_filename) do
        code = in_filename
        |> File.read!()
        |> run_trim()
        pre = File.read!("pre.html")
        |> String.replace(~S"#{NaiveDateTime.utc_now}", "#{NaiveDateTime.utc_now}")
        post = File.read!("post.html")
        File.write(out_filename, "#{pre}#{code}#{post}")
      end