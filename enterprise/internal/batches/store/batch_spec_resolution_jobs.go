package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// batchSpecResolutionJobInsertColumns is the list of changeset_jobs columns that are
// modified in CreateChangesetJob.
var batchSpecResolutionJobInsertColumns = []string{
	"batch_spec_id",
	"allow_unsupported",
	"allow_ignored",

	"state",

	"created_at",
	"updated_at",
}

// ChangesetJobColumns are used by the changeset job related Store methods to query
// and create changeset jobs.
var BatchSpecResolutionJobColums = SQLColumns{
	"batch_spec_resolution_jobs.id",

	"batch_spec_resolution_jobs.batch_spec_id",
	"batch_spec_resolution_jobs.allow_unsupported",
	"batch_spec_resolution_jobs.allow_ignored",

	"batch_spec_resolution_jobs.state",
	"batch_spec_resolution_jobs.failure_message",
	"batch_spec_resolution_jobs.started_at",
	"batch_spec_resolution_jobs.finished_at",
	"batch_spec_resolution_jobs.process_after",
	"batch_spec_resolution_jobs.num_resets",
	"batch_spec_resolution_jobs.num_failures",
	"batch_spec_resolution_jobs.execution_logs",
	"batch_spec_resolution_jobs.worker_hostname",

	"batch_spec_resolution_jobs.created_at",
	"batch_spec_resolution_jobs.updated_at",
}

// CreateBatchSpecResolutionJob creates the given batch spec resolutionjob jobs.
func (s *Store) CreateBatchSpecResolutionJob(ctx context.Context, ws ...*btypes.BatchSpecResolutionJob) (err error) {
	ctx, endObservation := s.operations.createBatchSpecResolutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("count", len(ws)),
	}})
	defer endObservation(1, observation.Args{})

	inserter := func(inserter *batch.Inserter) error {
		for _, wj := range ws {
			if wj.CreatedAt.IsZero() {
				wj.CreatedAt = s.now()
			}

			if wj.UpdatedAt.IsZero() {
				wj.UpdatedAt = wj.CreatedAt
			}

			state := string(wj.State)
			if state == "" {
				state = string(btypes.BatchSpecResolutionJobStateQueued)
			}

			if err := inserter.Insert(
				ctx,
				wj.BatchSpecID,
				wj.AllowUnsupported,
				wj.AllowIgnored,
				state,
				wj.CreatedAt,
				wj.UpdatedAt,
			); err != nil {
				return err
			}
		}

		return nil
	}
	i := -1
	return batch.WithInserterWithReturn(
		ctx,
		s.Handle().DB(),
		"batch_spec_resolution_jobs",
		batchSpecResolutionJobInsertColumns,
		BatchSpecResolutionJobColums,
		func(rows *sql.Rows) error {
			i++
			return scanBatchSpecResolutionJob(ws[i], rows)
		},
		inserter,
	)
}

// GetBatchSpecResolutionJobOpts captures the query options needed for getting a BatchSpecResolutionJob
type GetBatchSpecResolutionJobOpts struct {
	ID          int64
	BatchSpecID int64
}

// GetBatchSpecResolutionJob gets a BatchSpecResolutionJob matching the given options.
func (s *Store) GetBatchSpecResolutionJob(ctx context.Context, opts GetBatchSpecResolutionJobOpts) (job *btypes.BatchSpecResolutionJob, err error) {
	ctx, endObservation := s.operations.getBatchSpecResolutionJob.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("ID", int(opts.ID)),
		log.Int("BatchSpecID", int(opts.BatchSpecID)),
	}})
	defer endObservation(1, observation.Args{})

	q := getBatchSpecResolutionJobQuery(&opts)
	var c btypes.BatchSpecResolutionJob
	err = s.query(ctx, q, func(sc scanner) (err error) {
		return scanBatchSpecResolutionJob(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getBatchSpecResolutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_resolution_job.go:GetBatchSpecResolutionJob
SELECT %s FROM batch_spec_resolution_jobs
WHERE %s
LIMIT 1
`

func getBatchSpecResolutionJobQuery(opts *GetBatchSpecResolutionJobOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.id = %s", opts.ID))
	}

	if opts.BatchSpecID != 0 {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.batch_spec_id = %s", opts.BatchSpecID))
	}

	return sqlf.Sprintf(
		getBatchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListBatchSpecResolutionJobsOpts captures the query options needed for
// listing batch spec resolutionjob jobs.
type ListBatchSpecResolutionJobsOpts struct {
	State          btypes.BatchSpecResolutionJobState
	WorkerHostname string
}

// ListBatchSpecResolutionJobs lists batch changes with the given filters.
func (s *Store) ListBatchSpecResolutionJobs(ctx context.Context, opts ListBatchSpecResolutionJobsOpts) (cs []*btypes.BatchSpecResolutionJob, err error) {
	ctx, endObservation := s.operations.listBatchSpecResolutionJobs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	q := listBatchSpecResolutionJobsQuery(opts)

	cs = make([]*btypes.BatchSpecResolutionJob, 0)
	err = s.query(ctx, q, func(sc scanner) error {
		var c btypes.BatchSpecResolutionJob
		if err := scanBatchSpecResolutionJob(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	})

	return cs, err
}

var listBatchSpecResolutionJobsQueryFmtstr = `
-- source: enterprise/internal/batches/store/batch_spec_resolutionjob_job.go:ListBatchSpecResolutionJobs
SELECT %s FROM batch_spec_resolution_jobs
WHERE %s
ORDER BY id ASC
`

func listBatchSpecResolutionJobsQuery(opts ListBatchSpecResolutionJobsOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.State != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.state = %s", opts.State))
	}

	if opts.WorkerHostname != "" {
		preds = append(preds, sqlf.Sprintf("batch_spec_resolution_jobs.worker_hostname = %s", opts.WorkerHostname))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listBatchSpecResolutionJobsQueryFmtstr,
		sqlf.Join(BatchSpecResolutionJobColums.ToSqlf(), ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanBatchSpecResolutionJob(rj *btypes.BatchSpecResolutionJob, s scanner) error {
	var executionLogs []dbworkerstore.ExecutionLogEntry
	var failureMessage string

	if err := s.Scan(
		&rj.ID,
		&rj.BatchSpecID,
		&rj.AllowUnsupported,
		&rj.AllowIgnored,
		&rj.State,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &rj.StartedAt},
		&dbutil.NullTime{Time: &rj.FinishedAt},
		&dbutil.NullTime{Time: &rj.ProcessAfter},
		&rj.NumResets,
		&rj.NumFailures,
		pq.Array(&executionLogs),
		&rj.WorkerHostname,
		&rj.CreatedAt,
		&rj.UpdatedAt,
	); err != nil {
		return err
	}

	if failureMessage != "" {
		rj.FailureMessage = &failureMessage
	}

	for _, entry := range executionLogs {
		rj.ExecutionLogs = append(rj.ExecutionLogs, workerutil.ExecutionLogEntry(entry))
	}

	return nil
}

func ScanFirstBatchSpecResolutionJob(rows *sql.Rows, err error) (*btypes.BatchSpecResolutionJob, bool, error) {
	jobs, err := scanBatchSpecResolutionJobs(rows, err)
	if err != nil || len(jobs) == 0 {
		return nil, false, err
	}
	return jobs[0], true, nil
}

func scanBatchSpecResolutionJobs(rows *sql.Rows, queryErr error) ([]*btypes.BatchSpecResolutionJob, error) {
	if queryErr != nil {
		return nil, queryErr
	}

	var jobs []*btypes.BatchSpecResolutionJob

	return jobs, scanAll(rows, func(sc scanner) (err error) {
		var j btypes.BatchSpecResolutionJob
		if err = scanBatchSpecResolutionJob(&j, sc); err != nil {
			return err
		}
		jobs = append(jobs, &j)
		return nil
	})
}
