SELECT service, metric,
       toStartOfInterval(timestamp, toIntervalSecond(?)) AS ts,
       sum(value) AS units
FROM metrics FINAL
WHERE timestamp >= ? AND timestamp < ?
GROUP BY service, metric, ts
ORDER BY service, metric, ts;
