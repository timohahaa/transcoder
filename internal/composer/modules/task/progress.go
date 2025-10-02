package task

const (
	ProgressAfterDownloadSource = 7
	ProgressAfterUnmux          = 13
	ProgressAfterSplit          = 18
	ProgressAfterCreateSubtasks = 20

	progressAfterSplitting = 20
	progressAfterEncoding  = 80
	encodingPercentRange   = progressAfterEncoding - progressAfterSplitting

	ProgressAfterStitch        = 87
	ProgressAfterFragmentVideo = 95
	ProgressAfterFragmentAudio = 100
)
