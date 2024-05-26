local function rawtext (content_parts)
  local raw = ""
  for _, t in ipairs(content_parts) do
    if t.t == "Space" then
      raw = raw .. " "
    elseif t.t == "Str" then
      raw = raw .. t.text
    end
  end
  return raw
end

return {
  {
    Image = function (el)
      return pandoc.RawInline("markdown", "")
    end,
    Link = function (el)
      if el.title == "wikilink" then
        local target = el.target:gsub("_", " ")
        local raw = rawtext(el.content)

        if target:lower() == raw:lower() then
          return pandoc.RawInline("markdown", "[[" .. raw .. "]]")
        elseif raw == "" then
          return pandoc.RawInline("markdown", "[[" .. target .. "]]")
        else
          return pandoc.RawInline("markdown", "[[" .. target .. "|" .. raw .. "]]")
        end
      end

      el.classes = nil
      el.title = nil

      return el
    end,
  }
}
