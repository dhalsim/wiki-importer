local function rawtext(content_parts)
  local raw = ""
  for _, t in ipairs(content_parts) do
    if t.t == "Space" or t.t == "SoftBreak" then
      raw = raw .. " "
    elseif t.t == "Str" then
      raw = raw .. t.text
    end
  end
  return raw
end

return {
  {
    Image = function(el)
      return pandoc.RawInline("asciidoc", "")
    end,
    Link = function(el)
      -- Always treat as wikilink unless it's an explicit http(s) URL
      if not el.target:match("^https?://") then
        local target = el.target:gsub("_", " ")
        local raw = rawtext(el.content)
        
        if target:lower() == raw:lower() then
          return pandoc.RawInline("asciidoc", "[[" .. raw .. "]]")
        elseif raw == "" then
          return pandoc.RawInline("asciidoc", "[[" .. target .. "]]")
        else
          return pandoc.RawInline("asciidoc", "[[" .. target .. "|" .. raw .. "]]")
        end
      end
      
      return el
    end,
  }
}
