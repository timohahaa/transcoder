package ffprobe

type Info struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

type Stream struct {
	Index              int               `json:"index"`
	ID                 string            `json:"id"`
	CodecName          string            `json:"codec_name"`
	CodecLongName      string            `json:"codec_long_name"`
	CodecType          string            `json:"codec_type"`
	CodecTimeBase      string            `json:"codec_time_base"`
	CodecTagString     string            `json:"codec_tag_string"`
	CodecTag           string            `json:"codec_tag"`
	RFrameRate         string            `json:"r_frame_rate"`
	AvgFrameRate       string            `json:"avg_frame_rate"`
	TimeBase           string            `json:"time_base"`
	StartPts           int               `json:"start_pts"`
	StartTime          string            `json:"start_time"`
	DurationTs         uint64            `json:"duration_ts"`
	Duration           float64           `json:"duration,string"`
	BitRate            int64             `json:"bit_rate,string"`
	BitsPerRawSample   string            `json:"bits_per_raw_sample"`
	NbFrames           string            `json:"nb_frames"`
	Disposition        StreamDisposition `json:"disposition"`
	Tags               StreamTags        `json:"tags"`
	SideDataList       []SideData        `json:"side_data_list"`
	Profile            string            `json:"profile"`
	Width              int               `json:"width"`
	Height             int               `json:"height"`
	HasBFrames         int               `json:"has_b_frames"`
	SampleAspectRatio  string            `json:"sample_aspect_ratio"`
	DisplayAspectRatio string            `json:"display_aspect_ratio"`
	PixFmt             string            `json:"pix_fmt"`
	Level              int               `json:"level"`
	ColorRange         string            `json:"color_range"`
	ColorSpace         string            `json:"color_space"`
	ColorTransfer      string            `json:"color_transfer"`
	ColorPrimaries     string            `json:"color_primaries"`
	SampleFmt          string            `json:"sample_fmt"`
	SampleRate         string            `json:"sample_rate"`
	Channels           int               `json:"channels"`
	ChannelLayout      string            `json:"channel_layout"`
	BitsPerSample      int               `json:"bits_per_sample"`
}

type StreamDisposition struct {
	Default         int `json:"default"`
	Dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	CleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
}

type StreamTags struct {
	Rotate       int    `json:"rotate,string"`
	CreationTime string `json:"creation_time"`
	Language     string `json:"language"`
	Title        string `json:"title"`
	Encoder      string `json:"encoder"`
	Location     string `json:"location"`
}

type SideData struct {
	Type     string `json:"side_data_type"`
	Rotation int    `json:"rotation"`
}

type Format struct {
	NbStreams      int     `json:"nb_streams"`
	NbPrograms     int     `json:"nb_programs"`
	ProbeScore     int     `json:"probe_score"`
	Filename       string  `json:"filename"`
	FormatName     string  `json:"format_name"`
	FormatLongName string  `json:"format_long_name"`
	StartTime      string  `json:"start_time"`
	Duration       float64 `json:"duration,string"`
	Size           int64   `json:"size,string"`
	BitRate        int64   `json:"bit_rate,string"`
	Tags           Tags    `json:"tags"`
}

type Tags struct {
	Rotate           string `json:"rotate"`
	Language         string `json:"language"`
	HandlerName      string `json:"handler_name"`
	Encoder          string `json:"encoder"`
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	CreationTime     string `json:"creation_time"`
}
