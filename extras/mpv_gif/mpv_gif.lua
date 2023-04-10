-- Create animated GIFs with mpv and ffmpeg
-- Usage: "g" to set start frame, "G" to set end frame, "Ctrl+g" to create MPEG4_GIF, "Ctrl+G" to create mp4 with sound.

-- Credits: This is largely inspired by https://gist.github.com/Ruin0x11/8fae0a9341b41015935f76f913b28d2a

local msg = require 'mp.msg'
local utils = require 'mp.utils'

local mpeg4_gif_filters = ""
local standard_filters = "setsar=1:1"

local start_time = -1
local end_time = -1

-- shell escape
local is_windows = package.config:sub(1, 1) == "\\"
local home_path = os.getenv("HOME") or os.getenv("USERPROFILE")

local function esc(s)
    if is_windows then
        return s:gsub("\"", "\"\"")
    else
        return s:gsub("'", "'\\''")
    end
end

local function filename_esc(s)
    if is_windows then
        return s:gsub("%?", "_")
    else
        return s
    end
end

local function wrap_param(s)
    if is_windows then
        return "\"" .. s .. "\""
    else
        return "'" .. s .. "'"
    end
end

-- split a string
local function split(s, delimiter)
    local result = {}
    local from = 1
    local delim_from, delim_to = string.find(s, delimiter, from)
    while delim_from do
        table.insert(result, string.sub(s, from, delim_from - 1))
        from = delim_to + 1
        delim_from, delim_to = string.find(s, delimiter, from)
    end
    table.insert(result, string.sub(s, from))
    return result
end

local function starts_with(str, start)
    return str:sub(1, #start) == start
end

local function construct_filter(in_filter, max_aspect)
    local filters = {}

    if mp.get_property("deinterlace") == "yes" then
        -- If we are using deinterlace, let ffmpeg do it too.
        table.insert(filters, "yadif=1")
    end

    -- These aspects are after filter / manual aspect ratio change, but before output scaling (for window).
    local width = mp.get_property_number("dwidth")
    local height = mp.get_property_number("dheight")

    if (width > max_aspect) and (height > max_aspect) then
        -- OK we have to scale it.
        if width < height then
            width, height = max_aspect, height * max_aspect / width
        else
            width, height = width * max_aspect / height, max_aspect
        end
    end
    if width % 2 == 1 then
        width = width + 1
    end
    if height % 2 == 1 then
        height = height + 1
    end
    table.insert(filters, string.format("scale=%s:%s", width, height))

    table.insert(filters, standard_filters)

    if in_filter ~= "" then
        table.insert(filters, in_filter)
    end

    return table.concat(filters, ",")
end

local function detect_output_file_path(containing_path, filename, ext)
    local filename_with_ext = filename .. "." .. ext
    local preferred_file_path = utils.join_path(containing_path, filename_with_ext)

    local file, err = io.open(preferred_file_path, "w")
    if file then
        io.close(file)
        return preferred_file_path
    end

    local error_msg = string.format("Failed to write to preferred output (%s), writing to $HOME.", err)
    mp.osd_message(error_msg)
    msg.info(error_msg)
    return utils.join_path(home_path, filename_with_ext)
end

local function detect_dvd_bd_prefix(containing_path)
    local splitter = package.config:sub(1, 1)
    local segments = split(containing_path, splitter)

    for i = #segments, 1, -1 do
        if segments[i] == "" or segments[i] == "." then
            table.remove(segments, i)
        end
    end

    if segments[#segments] == "STREAM" then
        -- BDMV
        for i = #segments, 1, -1 do
            local s = segments[i]
            if not (s == "STREAM" or s == "BDMV" or s == "BD_VIDEO") then
                return s .. "_"
            end
        end
    elseif segments[#segments] == "VIDEO_TS" then
        -- DVD
        return segments[#segments - 1] .. "_"
    end
    return ""
end

local function ends_with(str, ending)
    return ending == "" or str:sub(-#ending) == ending
end

local function construct_input_and_seeking_args(input_file_path)
    local args = {}

    local output_seek_arg = ""
    if start_time ~= -1 then
        -- Need to seek
        if ends_with(mp.get_property("filename"), "ts") then
            -- Likely MPEGTS. ffmpeg seeking may not pick the correct key frame to start with.
            -- Let's do 2 seeks (-ss). The first one before input file (-i) to be start_time - 10.5 seconds.
            -- Hopefully we get a key frame. (x264's default key frame interval is 250 frames, at 24 fps it's 10.4s.)
            local input_seek_time = start_time - 10.5
            if input_seek_time < 0 then
                input_seek_time = 0
            end
            table.insert(args, string.format("-ss %s", input_seek_time))
            output_seek_arg = string.format("-ss %s", start_time - input_seek_time)
        else
            -- No need to do 2 seeks.
            table.insert(args, string.format("-ss %s", start_time))
        end
    end

    table.insert(args, string.format("-i %s", wrap_param(esc(input_file_path))))

    if output_seek_arg ~= "" then
        table.insert(args, output_seek_arg)
    end

    if end_time ~= -1 then
        local start_time_l = start_time
        if start_time_l == -1 then
            start_time_l = 0
        end
        table.insert(args, string.format("-t %s", end_time - start_time_l))
    end

    return table.concat(args, " ")
end

local function detect_edl_path(stream_path)
    if starts_with(stream_path, "edl://") then
        stream_path = stream_path:sub(7)
    end

    -- This is actually not correct, but YouTube URLs never contain this so...
    local segments = split(stream_path, ";")

    -- TODO: Handle multiple streams.
    for i, segment in ipairs(segments) do
        if (not starts_with(segment, "!")) and (not string.find(segment, "mime=audio")) then
            -- This is a file path, and if it's YouTube, mime is not audio/*.
            if starts_with(segment, "%") then
                -- This is the EDL style escape for comma.
                local len_str = string.match(segment, "%%(%d+)%%")
                local len = tonumber(len_str)
                return segment:sub(2 + #len_str + 1, 2 + #len_str + 1 + len - 1)
            else
                return split(segment, ",")[1]
            end
        end
    end
end

local function make_gif_internal(use_mpeg4)
    if start_time ~= -1 and end_time ~= -1 and start_time >= end_time then
        mp.osd_message("Invalid start/end time.")
        return
    end

    mp.osd_message("Creating GIF.")

    local input_file_path = ""
    local stream_path = mp.get_property("stream-open-filename")
    if starts_with(stream_path, "edl://") then
        -- EDL. Likely ytdl or something. Attempt to parse it ourselves.
        input_file_path = detect_edl_path(stream_path)
    elseif string.find(stream_path, "://") then
        -- Likely some network source.
        input_file_path = stream_path
    else
        input_file_path = utils.join_path(mp.get_property("working-directory"), mp.get_property("path"))
    end

    local input_and_seeking_args = construct_input_and_seeking_args(input_file_path)

    local containing_path = utils.split_path(input_file_path)
    local input_file_name_no_ext = mp.get_property("filename/no-ext")

    local output_filename = detect_dvd_bd_prefix(containing_path) .. string.format('%s_%s_%s', filename_esc(input_file_name_no_ext), start_time, end_time)
    local output_file_path

    if use_mpeg4 then
        -- MPEG4_GIF
        output_file_path = detect_output_file_path(containing_path, output_filename, "mp4")

        local filters = construct_filter(mpeg4_gif_filters, 720)

        local args = string.format("ffmpeg -v warning %s -map_chapters -1 -c:v libx264 -pix_fmt yuv420p -an -filter:v %s -y %s", input_and_seeking_args, wrap_param(filters), wrap_param(esc(output_file_path)))
        msg.info(args)
        os.execute(args)
    else
        -- MP4
        output_file_path = detect_output_file_path(containing_path, output_filename, "a.mp4")

        local filters = construct_filter(mpeg4_gif_filters, 1080)

        local args = string.format("ffmpeg -v warning %s -map_chapters -1 -c:v libx264 -pix_fmt yuv420p -filter:v %s -y %s", input_and_seeking_args, wrap_param(filters), wrap_param(esc(output_file_path)))
        msg.info(args)
        os.execute(args)
    end

    msg.info("GIF created.")
    mp.osd_message(string.format("GIF created.\n%s", output_file_path))
end

local function set_gif_start()
    start_time = mp.get_property_number("time-pos", -1)
    mp.osd_message("GIF Start: " .. start_time)
end

local function set_gif_end()
    end_time = mp.get_property_number("time-pos", -1)
    mp.osd_message("GIF End: " .. end_time)
end

local function make_mpeg4_gif()
    make_gif_internal(true)
end

local function make_gif()
    make_gif_internal(false)
end

mp.add_key_binding("g", "set_gif_start", set_gif_start)
mp.add_key_binding("G", "set_gif_end", set_gif_end)
mp.add_key_binding("Ctrl+g", "make_mpeg4_gif", make_mpeg4_gif)
mp.add_key_binding("Ctrl+G", "make_gif", make_gif)

local function toggle_mpdecimate()
    mp.command("vf toggle @mpvgifmpdecimate")
end

mp.commandv("vf", "add", "@mpvgifmpdecimate:!lavfi=mpdecimate")
mp.add_key_binding("D", "toggle_mpdecimate", toggle_mpdecimate)
