-- Create animated GIFs with mpv and ffmpeg
-- Usage: "g" to set start frame, "G" to set end frame, "Ctrl+g" to create MPEG4_GIF, "Ctrl+G" to create GIF.
local msg = require 'mp.msg'
local utils = require 'mp.utils'

gif_filters = "fps=24"
mpeg4_gif_filters = ""
standard_filters = "setsar=1:1"

start_time = -1
end_time = -1
temp_palette_path = "/tmp/palette.png"

-- shell escape
function esc(s)
    return string.gsub(s, "'", "'\\''")
end

function make_mpeg4_gif()
    make_gif_internal(true)
end

function make_gif()
    make_gif_internal(false)
end

function construct_filter(in_filter, max_aspect)
    filters = {}

    if mp.get_property("deinterlace") == "yes" then
        -- If we are using deinterlace, let ffmpeg do it too.
        table.insert(filters, "yadif")
    end

    -- These aspects are after filter / manual aspect ratio change, but before output scaling (for window).
    width = mp.get_property_number("dwidth")
    height = mp.get_property_number("dheight")

    if (width > max_aspect) and (height > max_aspect) then
        -- OK we have to scale it.
        if width < height then
            table.insert(filters, string.format("scale=%s:%s", max_aspect, height * max_aspect / width))
        else
            table.insert(filters, string.format("scale=%s:%s", width * max_aspect / height, max_aspect))
        end
    end

    table.insert(filters, standard_filters)

    if in_filter ~= "" then
        table.insert(filters, in_filter)
    end

    return table.concat(filters, ",")
end

function make_gif_internal(use_mpeg4)
    local start_time_l = start_time
    local end_time_l = end_time
    if start_time_l == -1 or end_time_l == -1 or start_time_l >= end_time_l then
        mp.osd_message("Invalid start/end time.")
        return
    end

    local position = start_time_l
    local duration = end_time_l - start_time_l

    mp.osd_message("Creating GIF.")

    local input_file_path = utils.join_path(mp.get_property("working-directory"), mp.get_property("path"))
    local containing_path = utils.split_path(input_file_path)
    local input_file_name_no_ext = mp.get_property("filename/no-ext")

    local output_file_path = utils.join_path(containing_path, string.format('%s_%s_%s', input_file_name_no_ext, start_time_l, end_time_l))

    if use_mpeg4 then
        -- MPEG4_GIF
        output_file_path = output_file_path .. ".mp4"

        filters = construct_filter(mpeg4_gif_filters, 720)

        local args = string.format("ffmpeg -v warning -ss %s -i '%s' -t %s -c:v libx264 -pix_fmt yuv420p -an -filter:v \"%s\" -y '%s'", position, esc(input_file_path), duration, filters, esc(output_file_path))
        msg.info(args)
        os.execute(args)
    else
        -- Real GIF
        output_file_path = output_file_path .. ".gif"

        filters = construct_filter(gif_filters, 540)

        -- first, create the palette
        local args = string.format("ffmpeg -v warning -ss %s -t %s -i '%s' -vf '%s,palettegen' -y '%s'", position, duration, esc(input_file_path), filters, esc(temp_palette_path))
        msg.info(args)
        os.execute(args)

        -- then, create GIF
        args = string.format("ffmpeg -v warning -ss %s -t %s -i '%s' -i '%s' -lavfi '%s [x]; [x][1:v] paletteuse' -y '%s'", position, duration, esc(input_file_path), esc(temp_palette_path), filters, esc(output_file_path))
        msg.info(args)
        os.execute(args)
    end

    msg.info("GIF created.")
    mp.osd_message(string.format("GIF created.\n%s", output_file_path))
end

function set_gif_start()
    start_time = mp.get_property_number("time-pos", -1)
    mp.osd_message("GIF Start: " .. start_time)
end

function set_gif_end()
    end_time = mp.get_property_number("time-pos", -1)
    mp.osd_message("GIF End: " .. end_time)
end

mp.add_key_binding("g", "set_gif_start", set_gif_start)
mp.add_key_binding("G", "set_gif_end", set_gif_end)
mp.add_key_binding("Ctrl+g", "make_gif", make_mpeg4_gif)
mp.add_key_binding("Ctrl+G", "make_gif_with_subtitles", make_gif)
