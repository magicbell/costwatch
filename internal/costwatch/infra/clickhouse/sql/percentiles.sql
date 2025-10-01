WITH
	? AS start_ts,
	? AS end_ts,
	toIntervalSecond (?) AS bucket,
	toStartOfInterval (end_ts, bucket) AS end_bucket,
	by_bucket AS (
		SELECT
			service,
			metric,
			toStartOfInterval (timestamp, bucket) AS bucket_ts,
			sum(value) AS bucket_usage
		FROM
			metrics FINAL
		WHERE
			timestamp >= start_ts
			AND timestamp < end_bucket
		GROUP BY
			service,
			metric,
			bucket_ts
	)
SELECT
	service,
	metric,
	toFloat64 (quantileTDigest (0.50) (bucket_usage)) AS p50,
	toFloat64 (quantileTDigest (0.90) (bucket_usage)) AS p90,
	toFloat64 (quantileTDigest (0.95) (bucket_usage)) AS p95,
	toFloat64 (max(bucket_usage)) AS pmax
FROM
	by_bucket
WHERE
	service NOT IN (?)
GROUP BY
	service,
	metric
ORDER BY
	service,
	metric;