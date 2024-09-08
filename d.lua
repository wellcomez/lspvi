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
local function get_available_themes()
    local themes = {}

    -- Get runtime path
    local runtimepath = vim.o.runtimepath
    for path in runtimepath:gmatch("[^,]+") do
        local colors_dir = path .. "/colors"
        -- Check if the colors directory exists
        if vim.fn.isdirectory(colors_dir) == 1 then
            -- List all .vim files in the colors directory
            for _, file in ipairs(vim.fn.split(vim.fn.globpath(colors_dir, "*.vim"), "\n")) do
                local theme_name = vim.fn.fnamemodify(file, ":t:r") -- Get filename without extension
                if theme_name and not vim.tbl_contains(themes, theme_name) then
                    table.insert(themes, theme_name)
                end
            end
        end
    end

    return themes
end
local colorscheme = vim.g.colors_name
local function save_scheme(colorscheme)
    local output_file = colorscheme .. ".yml" -- Path to the output file
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
            f:write(" - Group: " .. '"' .. group .. '"' .. "\n")
            for k, v in pairs(hl) do
                if k == "foreground" or k == "background" then
                    local d = string.format("#%x", v)
                    f:write("   " .. tostring(k) .. ": " .. tostring(d) .. "\n")
                else
                    f:write("   " .. tostring(k) .. ": " .. tostring(v) .. "\n")
                end
            end
            f:write("\n")
        end

        -- Close the file
        f:close()

        print("Highlight settings saved to " .. file)
    end

    -- Call the function
    save_highlights_to_file(highlight_groups, output_file)
end
save_scheme(colorscheme)
--TSVariable
