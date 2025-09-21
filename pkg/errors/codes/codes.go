package codes

const (
	Unknown           = "UNKNOWN_ERROR"
	ChunkOverflow     = "CHUNK_OVERFLOW"
	TaskResetByWorker = "TASK_RESET_BY_WORKER"
	Generic           = "GENERIC_ERROR"
	Network           = "NETWORK_ERROR"
	Ffmpeg            = "FFMPEG_ERROR"
	SplitSources      = "SPLIT_SOURCES_ERROR"
	StitchSources     = "STITCH_SOURCES_ERROR"
	FragmentSources   = "FRAGMENT_SOURCES_ERROR"
	EncryptSources    = "ENCRYPT_SOURCES_ERROR"
	GeneratePoster    = "GENERATE_POSTER_ERROR"
	DB                = "DB_ERROR"
	Redis             = "REDIS_ERROR"
	Upload            = "UPLOAD_ERROR"
	UnmuxAudio        = "UNMUX_AUDIO_ERROR"
	Validation        = "VALIDATION"
)
