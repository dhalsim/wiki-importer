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
  io.stderr:write("\n=== DefinitionList Debug ===\n")
  io.stderr:write("Full element: " .. dump(el) .. "\n")

  -- Create an array to hold all processed items
  local result = {}
  
  -- Process each item in the definition list
  for _, item in ipairs(el.content) do
    -- Process all definitions in this item
    if item[2] then
      -- Check if this is a first pass or second pass structure
      local def = item[2][1]
      io.stderr:write("\nChecking def structure:\n")
      io.stderr:write("def type: " .. type(def) .. "\n")
      io.stderr:write("def content: " .. dump(def) .. "\n")
      
      -- Check if this is a first pass (def[1] is a Plain block)
      local is_first_pass = def[1] and def[1].t == "Plain"
      io.stderr:write("Is first pass: " .. tostring(is_first_pass) .. "\n")
      
      if is_first_pass then
        -- First pass: Process each definition
        for _, d in ipairs(item[2]) do
          local plain = d[1]
          if plain.t == "Plain" then
            -- Skip if already processed
            if #plain.content == 1 and plain.content[1].t == "RawInline" then
              table.insert(result, plain)
            else
              -- Process inline elements
              local content = ""
              for _, inline in ipairs(plain.content) do
                if inline.t == "Str" then
                  content = content .. inline.text
                elseif inline.t == "Space" then
                  content = content .. " "
                elseif inline.t == "Link" then
                  if inline.attr and inline.attr.classes and inline.attr.classes[1] == "wikilink" then
                    content = content .. "[[" .. pandoc.utils.stringify(inline.content) .. "]]"
                  else
                    content = content .. pandoc.utils.stringify(inline.content)
                  end
                end
              end
              
              -- Create a RawBlock with newline
              table.insert(result, pandoc.RawBlock("asciidoc", ":: " .. content .. "\n"))
            end
          end
        end
      else
        -- Second pass: All RawBlocks are in the first definition
        io.stderr:write("\nSecond pass processing:\n")
        io.stderr:write("Number of items in def: " .. #def .. "\n")
        for i = 1, #def do
          local block = def[i]
          io.stderr:write("Processing item " .. i .. ": " .. dump(block) .. "\n")
          if block.t == "RawBlock" then
            io.stderr:write("Adding RawBlock to result\n")
            -- Add newline if it's not already there
            if not block.text:match("\n$") then
              block.text = block.text .. "\n"
            end
            table.insert(result, block)
          end
        end
      end
    end
  end
  
  io.stderr:write("\nFinal result: " .. dump(result) .. "\n")
  -- Remove the last newline to match the expected output exactly
  if #result > 0 then
    result[#result].text = result[#result].text:gsub("\n$", "")
  end
  return result  -- Return array of RawBlocks
end