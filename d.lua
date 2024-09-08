-- Define highlight groups and output file path
---@type
local highlight_groups = {
    "Comment",
    --"Variable",
    "Constant",
    "String",
    "Function",
    "Keyword",
    "Type",
    "Character",
    "Number",
    "Boolean",
    "Float",
    "Identifier",
    "Function",
    "Statement",
    "Conditional",
    "Repeat",
    "Label",
    "Operator",
    "Keyword",
    "Exception",
    "PreProc",
    "Include",
    "Define",
    "Macro",
    "PreCondit",
    "Type",
    "StorageClass",
    "Structure",
    "Typedef",
    "Special",
    "SpecialChar",
    "Tag",
    "Delimiter",
    "SpecialComment",
    "Debug",
    "Underlined",
    "Ignore",
    "Error",
    "Added",
    "Changed",
    "Removed",
    "CursorLine",
    "CursorColumn",
    "Visual",
    "StatusLine",
    "Normal",
    "DiagnosticError",
    "DiagnosticWarn",
    "DiagnosticInfo",
    "DiagnosticHint",
    --> 'TSKeyword', 'TSFunction','TSMethod',
    "DiffAdd",
    "DiffChange",
    "DiffDelete",
    "LspReferenceText",
    "LspReferenceRead",
    "LspReferenceWrite",
    "@variable",
    "@variable.member",
    "@variable.parameter",
    "@function",
    "@function.method",
    "@function.call",
    "@keyword",
    "@keyword.type",
    "@keyword.Conditional",
    "@keyword.function",
    "@keyword.return",
    "@keyword.modifier",
    "@keyword.import",
    "@module",
    "@String",
    "@type",
    "@type.definition",
    "@type.class",
    "@property",
    "@Operator",
    "@Boolean",
    "@Comment",
    "@Constant",
    "@Constant.builtin"
}

local output_file = "all_highlight_groups_info.yml" -- Path to the output file

-- Function to get highlight settings and write to a file
local function save_highlights_to_file(groups, file)
    -- Open the file for writing
    local f = io.open(file, "w")
    if not f then
        print("Error opening file!")
        return
    end

    -- Write the highlight settings to the file
      f:write("Data:\n")
    for _, group in ipairs(groups) do
        local hl = vim.api.nvim_get_hl_by_name(group, true)
        f:write(" - Group: " .. group .. "\n")
        for k, v in pairs(hl) do
            f:write("   "..tostring(k) .. ": " .. tostring(v) .. "\n")
        end
        f:write("\n")
    end

    -- Close the file
    f:close()

    print("Highlight settings saved to " .. file)
end

-- Call the function
save_highlights_to_file(highlight_groups, output_file)
--TSVariable
