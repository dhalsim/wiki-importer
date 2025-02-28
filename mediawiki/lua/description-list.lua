-- Utility function to help with debugging
-- Converts any Lua value into a string representation
function dump(o)
  if type(o) == 'table' then
    local s = '{ '
    for k,v in pairs(o) do
      if type(k) ~= 'number' then k = '"'..k..'"' end
      s = s .. '['..k..'] = ' .. dump(v) .. ','
    end
    return s .. '} '
  else
    return tostring(o)
  end
end

-- Helper function to convert inline elements to string
-- Handles Str (text), Space, and Link elements
function Inlines_to_string(inlines)
  local result = ""
  for _, inline in pairs(inlines) do
    if inline.t == "Str" then
      result = result .. inline.text
    elseif inline.t == "Space" then
      result = result .. " "
    elseif inline.t == "Link" and inline.attr.classes[1] == "wikilink" then
      result = result .. "[[" .. pandoc.utils.stringify(inline.content) .. "]]"
    end
  end
  return result
end

-- Create a test definition list to see its structure
function create_test_deflist()
  -- Create a term (empty in MediaWiki's :: case)
  local term = {}
  
  -- Create the definition content
  local str1 = pandoc.Str("Ellora")
  local space1 = pandoc.Space()
  local str2 = pandoc.Str("Section:")
  local space2 = pandoc.Space()
  local link = pandoc.Link("Serapis Bey", "Serapis_Bey", "", pandoc.Attr("", {"wikilink"}))
  
  -- Create a Plain block containing the inlines
  local plain = pandoc.Plain({str1, space1, str2, space2, link})
  
  -- Create the definition (a list containing the Plain block)
  local definition = {plain}
  
  -- Create the definition list item (a pair of term and definitions)
  local item = {term, {definition}}
  
  -- Create the definition list
  local deflist = pandoc.DefinitionList({item})
  
  io.stderr:write("\n=== Test DefinitionList Structure ===\n")
  io.stderr:write(dump(deflist) .. "\n")
  return deflist
end

-- Main filter function for DefinitionList elements
-- Converts MediaWiki's :: syntax to AsciiDoc format while preserving wikilinks
function DefinitionList(el)
  -- Extract the first definition's content from the nested structure:
  -- el.content[1] -> first item
  -- [2] -> definitions part (array)
  -- [1] -> first definition
  -- [1] -> first block in definition
  if el.content and el.content[1] and el.content[1][2] and el.content[1][2][1] and el.content[1][2][1][1] then
    local plain = el.content[1][2][1][1]
    
    -- Only process Plain blocks (not other block types)
    if plain.t == "Plain" then
      -- Skip if already processed (contains RawInline)
      if #plain.content == 1 and plain.content[1].t == "RawInline" then
        return plain
      end
      
      -- Process each inline element and build the content string
      local content = ""
      for _, inline in ipairs(plain.content) do
        if inline.t == "Str" then
          -- Regular text
          content = content .. inline.text
        elseif inline.t == "Space" then
          -- Space between elements
          content = content .. " "
        elseif inline.t == "Link" then
          -- Handle links: preserve wikilinks, convert others to text
          if inline.attr and inline.attr.classes and inline.attr.classes[1] == "wikilink" then
            content = content .. "[[" .. pandoc.utils.stringify(inline.content) .. "]]"
          else
            content = content .. pandoc.utils.stringify(inline.content)
          end
        end
      end
      
      -- Create a Plain block containing a RawInline with AsciiDoc format
      return pandoc.Plain({
        pandoc.RawInline("asciidoc", ":: " .. content)
      })
    end
  end
  
  -- Return nil if we can't process this element
  -- (tells pandoc to leave it unchanged)
  return nil
end