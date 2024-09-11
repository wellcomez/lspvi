-- Define the highlight group and output file path
local highlight_group = 'Comment' -- Replace with the highlight group you want
local output_file = '/Users/jialaizhu/highlight_info.txt' -- Path to the output file

-- Function to get highlight settings and write to a file
local function save_highlight_to_file(group, file)
    -- Get the highlight settings
    local hl = vim.api.nvim_get_hl_by_name(group, true)
    
    -- Open the file for writing
    local f = io.open(file, 'w')
    if not f then
        print("Error opening file!")
        return
    end
    
    -- Write the highlight settings to the file
    f:write("Highlight Group: " .. group .. "\n")
    for k, v in pairs(hl) do
        f:write(k .. ": " .. tostring(v) .. "\n")
    end
    
    -- Close the file
    f:close()
    
    print("Highlight settings saved to " .. file)
end

-- Call the function
save_highlight_to_file(highlight_group, output_file)

