-- Define highlight groups and output file path
---@type
local highlight_groups = {"Comment", -- "Variable",
  "Constant", "String", "Function", "Keyword", "Type", "Character", "Number", "Boolean", "Float", "Identifier",
  "Function", "Statement", "Conditional", "Repeat", "Label", "Operator", "Keyword", "Exception",
  "PreProc", "Include", "Define", "Macro", "PreCondit", "Type", "StorageClass", "Structure",
  "Typedef", "Special", "SpecialChar", "Tag", "Delimiter", "SpecialComment", "Debug",
  "Underlined", "Ignore", "Error", "Added", "Changed", "Removed", "CursorLine", "CursorColumn",
  "Visual", "StatusLine", "Normal", "DiagnosticError", "DiagnosticWarn", "DiagnosticInfo",
  "DiagnosticHint", -- > 'TSKeyword', 'TSFunction','TSMethod',
  "DiffAdd", "DiffChange", "DiffDelete", "LspReferenceText", "LspReferenceRead", "LspReferenceWrite", "@variable", --                    ; various variable names
  "@variable.builtin", --            ; built-in variable names (e.g. `this`)
  "@variable.parameter", --          ; parameters of a function
  "@variable.parameter", -- .builtin  ; special parameters (e.g. `_`, `it`)
  "@variable.member", --             ; object and struct fields
  "@constant", --          ; constant identifiers
  "@constant.builtin", --  ; built-in constant values
  "@constant.macro", --    ; constants defined by the preprocessor
  "@module", --            ; modules or namespaces
  "@module.builtin", --    ; built-in modules or namespaces
  "@label", --             ; GOTO and other labels (e.g. `label:` in C), including heredoc labels
  "@function", --             ; function definitions
  "@function.builtin", --     ; built-in functions
  "@function.call", --        ; function calls
  "@function.macro", --       ; preprocessor macros
  "@function.method", --      ; method definitions
  "@function.method", -- .call ; method calls
  "@constructor", --          ; constructor calls and definitions
  "@operator", --         ; symbolic operators (e.g. `+` / `*`)
  "@String", "@type", --             ; type or class definitions and annotations
  "@type.builtin", --     ; built-in types
  "@type.definition", --  ; identifiers in type definitions (e.g. `typedef <type> <identifier>` in C)
  "@type.class", "@attribute", --          ; attribute annotations (e.g. Python decorators, Rust lifetimes)
  "@attribute.builtin", --  ; builtin annotations (e.g. `@property` in Python)
  "@property", --           ; the key in key/value pairs
  "@Comment", 
  "@keyword", "@keyword.coroutine", "@keyword.function", "@keyword.operator", "@keyword.import",
  "@keyword.type", "@keyword.modifier", "@keyword.repeat", "@keyword.return", "@keyword.debug",
  "@keyword.exception", "@keyword.conditional", "@keyword.conditional.ternary",
  "@keyword.directive", "@keyword.directive.define", 
  "@string", --                 ; string literals
  "@string.documentation", --   ; string documenting code (e.g. Python docstrings)
  "@string.regexp", --          ; regular expressions
  "@string.escape", --          ; escape sequences
  "@string.special", --         ; other special strings (e.g. dates)
  "@string.special", -- .symbol  ; symbols or atoms
  "@string.special", -- .url     ; URIs (e.g. hyperlinks)
  "@string.special", -- .path    ; filenames
  "@character", --              ; character literals
  "@character.special", --      ; special characters (e.g. wildcards)
  "@boolean", --                ; boolean literals
  "@number", --                 ; numeric literals
  "@number.float", --           ; floating-point number literals
  "@markup.strong", --         ; bold text
  "@markup.italic", --         ; italic text
  "@markup.strikethrough", --  ; struck-through text
  "@markup.underline", --      ; underlined text (only for literal underline markup!)
  "@markup.heading", --        ; headings, titles (including markers)
  "@markup.heading", -- .1      ; top-level heading
  "@markup.heading", -- .2      ; section heading
  "@markup.heading", -- .3      ; subsection heading
  "@markup.heading", -- .4      ; and so on
  "@markup.heading", -- .5      ; and so forth
  "@markup.heading", -- .6      ; six levels ought to be enough for anybody
  "@markup.quote", --          ; block quotes
  "@markup.math", --           ; math environments (e.g. `$ ... $` in LaTeX)
  "@markup.link", --           ; text references, footnotes, citations, etc.
  "@markup.link", -- .label     ; link, reference descriptions
  "@markup.link", -- .url       ; URL-style links
  "@markup.raw", --            ; literal or verbatim text (e.g. inline code)
  "@markup.raw", -- .block      ; literal or verbatim text as a stand-alone block
  --                       ; (use priority 90 for blocks with injections)
  "@markup.list", --           ; list markers
  "@markup.list", -- .checked   ; checked todo-style list markers
  "@markup.list" -- .unchecked ; unchecked todo-style list markers
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
                    f:write("   " .. tostring(k) .. ": \"" .. tostring(d) .. "\"\n")
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

local function set_colorscheme(theme_name)
    -- Validate that the theme exists
    local success, _ = pcall(vim.cmd, "colorscheme " .. theme_name)
    if not success then
        print("Colorscheme '" .. theme_name .. "' not found.")
    else
        print("Colorscheme set to '" .. theme_name .. "'.")
    end
end

local function save_themes_to_file(themes)
    for _, theme in ipairs(themes) do
        set_colorscheme(theme)
        save_scheme(theme)
    end
end
local themes = get_available_themes()
save_themes_to_file(themes)
-- save_scheme(colorscheme)
-- TSVariable
