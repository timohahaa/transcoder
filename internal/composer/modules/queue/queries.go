package queue

const (
	lockQuery = `SELECT pg_try_advisory_xact_lock($1)`

	checkTasksExistQuery = `
	SELECT 
		COALESCE(SUM(duration)  > $2, FALSE) OR 
		COALESCE(SUM(file_size) > $3, FALSE) OR 
		COUNT(*) > $4
	FROM transcoder.queue
	WHERE routing = $1
		AND status IN ('waiting-splitting', 'splitting')
		AND deleted_at IS NULL`

	selectTasksQuery = `
	SELECT
		task_id
	FROM transcoder.queue
	WHERE status = 'pending'
		AND deleted_at IS NULL
	ORDER BY created_at ASC`

	setWaitingSplittingQuery = `
	UPDATE transcoder.queue
	SET
		  status     = 'waiting-splitting'
		, routing    = $2
		, updated_at = CURRENT_TIMESTAMP
	WHERE task_id = $1`
)
