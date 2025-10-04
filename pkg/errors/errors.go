package errors

import (
	"strconv"

	"github.com/timohahaa/transcoder/pkg/errors/codes"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

func New(reason, domain string, metadata map[string]string) *pb.Error {
	return &pb.Error{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	}
}

func extractMeta(err error) map[string]string {
	var meta = make(map[string]string)
	switch e := err.(type) {
	case *ffmpeg.Error:
		meta["code"] = e.Code
		meta["message"] = e.Message
		meta["path"] = e.Path
		meta["timestamp"] = e.GetTimestamp()
	default:
		meta["err"] = err.Error()
	}
	return meta
}

func PreValidation(err error) *pb.Error {
	return New(codes.Validation, "splitter", extractMeta(err))
}

func PostValidation(err error) *pb.Error {
	return New(codes.Validation, "assembler", extractMeta(err))
}

func Unknown(err error) *pb.Error {
	return New(codes.Unknown, "transcoder", extractMeta(err))
}

func Generic(err error) *pb.Error {
	return New(codes.Generic, "transcoder", extractMeta(err))
}

func Splitter(err error) *pb.Error {
	return New(codes.Generic, "splitter", extractMeta(err))
}

func Assembler(err error) *pb.Error {
	return New(codes.Generic, "assembler", extractMeta(err))
}

func Encoder(err error) *pb.Error {
	return New(codes.Generic, "encoder", extractMeta(err))
}

func Network(err error) *pb.Error {
	return New(codes.Network, "transcoder", extractMeta(err))
}

func DB(err error) *pb.Error {
	return New(codes.DB, "transcoder", extractMeta(err))
}

func Redis(err error) *pb.Error {
	return New(codes.Redis, "transcoder", extractMeta(err))
}

func Upload(err error) *pb.Error {
	return New(codes.Upload, "transcoder", extractMeta(err))
}

func Ffmpeg(err error) *pb.Error {
	var ffmpegErr *ffmpeg.Error
	switch e := err.(type) {
	case *ffmpeg.Error:
		ffmpegErr = e
	default:
		ffmpegErr = &ffmpeg.Error{
			Code:    ffmpeg.ErrUnknown,
			Message: err.Error(),
		}
	}

	return New(codes.Network, "ffmpeg", map[string]string{
		"timestamp": ffmpegErr.GetTimestamp(),
		"code":      ffmpegErr.Code,
		"path":      ffmpegErr.Path,
		"message":   ffmpegErr.Message,
	})
}

func SplitSources(err error) *pb.Error {
	return New(codes.SplitSources, "splitter", extractMeta(err))
}

func StitchSources(err error) *pb.Error {
	return New(codes.StitchSources, "assembler", extractMeta(err))
}

func FragmentSources(err error) *pb.Error {
	return New(codes.FragmentSources, "assembler", extractMeta(err))
}

func EncryptSources(err error) *pb.Error {
	return New(codes.EncryptSources, "assembler", extractMeta(err))
}

func Unmux(err error) *pb.Error {
	return New(codes.Unmux, "splitter", extractMeta(err))
}

func TaskReset() *pb.Error {
	return New(codes.TaskResetByEncoder, "encoder", nil)
}

func GeneratePoster(err error) *pb.Error {
	return New(codes.GeneratePoster, "encoder", extractMeta(err))
}

func ChunkOverflow(need, actual int32) *pb.Error {
	return New(codes.ChunkOverflow, "composer", map[string]string{
		"need":   strconv.FormatInt(int64(need), 10),
		"actual": strconv.FormatInt(int64(actual), 10),
	})
}
