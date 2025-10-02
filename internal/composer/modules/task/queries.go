package task

const (
	getForSplittingQuery = `
	UPDATE transcoder.queue
	SET
		status     = 'splitting'
        , hostname = $1
		, updated_at = CURRENT_TIMESTAMP
	WHERE task_id = (
		SELECT
			task_id
		FROM transcoder.queue
		WHERE status = 'waiting-splitting'
			AND hostname IN ($1, '')
			AND deleted_at IS NULL
		ORDER BY LENGTH(hostname) DESC, task_id
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	) RETURNING
		task_id
		, source
		, encoder
		, routing
        , duration
        , file_size
        , settings
	`

	getForAssemblingQuery = `
	UPDATE transcoder.queue
	SET
		status     = 'assembling'
		, updated_at = CURRENT_TIMESTAMP
	WHERE task_id = (
		SELECT
			task_id
		FROM transcoder.queue
		WHERE status = 'waiting-assembling'
			AND deleted_at IS NULL
            AND hostname = $1
		ORDER BY task_id
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	) RETURNING
		task_id
		, source
		, encoder
		, routing
        , duration
        , file_size
        , settings
	`

	checkCancellationQuery = `
	SELECT
		(
			deleted_at IS NOT NULL
			OR status = 'canceled'
		)
	FROM transcoder.queue
	WHERE task_id = $1
	`

	updateTaskStatusQuery = `
    UPDATE transcoder.queue
    SET
        updated_at = CURRENT_TIMESTAMP
        , status = $2
        , error = $3
    WHERE task_id = $1
		AND status != 'done'
    `

	getTaskDurationQuery = `
	SELECT 
		duration
	FROM transcoder.queue
    WHERE task_id = $1
        AND deleted_at IS NULL
    `
)
