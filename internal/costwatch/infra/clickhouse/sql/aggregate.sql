SELECT
  service,
  metric,
  toStartOfInterval (timestamp, toIntervalSecond (?)) AS ts,
  sum(value) AS units
FROM
  metrics FINAL
WHERE
  timestamp >= ?
  AND timestamp < ?
  AND service NOT IN (?)
GROUP BY
  service,
  metric,
  ts
ORDER BY
  service,
  metric,
  ts;