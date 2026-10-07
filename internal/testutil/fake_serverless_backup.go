package testutil

import (
	"context"

	spacebackupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/serverless/space/backup/v1"
)

// FakeServerlessBackupService is a test fake that implements the serverless BackupServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeServerlessBackupService struct {
	spacebackupv1.UnimplementedBackupServiceServer

	ListBackupsCalls          MethodSpy[*spacebackupv1.ListBackupsRequest, *spacebackupv1.ListBackupsResponse]
	GetBackupCalls            MethodSpy[*spacebackupv1.GetBackupRequest, *spacebackupv1.GetBackupResponse]
	CreateBackupCalls         MethodSpy[*spacebackupv1.CreateBackupRequest, *spacebackupv1.CreateBackupResponse]
	DeleteBackupCalls         MethodSpy[*spacebackupv1.DeleteBackupRequest, *spacebackupv1.DeleteBackupResponse]
	ListBackupRestoresCalls   MethodSpy[*spacebackupv1.ListBackupRestoresRequest, *spacebackupv1.ListBackupRestoresResponse]
	RestoreBackupCalls        MethodSpy[*spacebackupv1.RestoreBackupRequest, *spacebackupv1.RestoreBackupResponse]
	ListBackupSchedulesCalls  MethodSpy[*spacebackupv1.ListBackupSchedulesRequest, *spacebackupv1.ListBackupSchedulesResponse]
	GetBackupScheduleCalls    MethodSpy[*spacebackupv1.GetBackupScheduleRequest, *spacebackupv1.GetBackupScheduleResponse]
	CreateBackupScheduleCalls MethodSpy[*spacebackupv1.CreateBackupScheduleRequest, *spacebackupv1.CreateBackupScheduleResponse]
	UpdateBackupScheduleCalls MethodSpy[*spacebackupv1.UpdateBackupScheduleRequest, *spacebackupv1.UpdateBackupScheduleResponse]
	DeleteBackupScheduleCalls MethodSpy[*spacebackupv1.DeleteBackupScheduleRequest, *spacebackupv1.DeleteBackupScheduleResponse]
}

// ListBackups records the call and dispatches via ListBackupsCalls.
func (f *FakeServerlessBackupService) ListBackups(ctx context.Context, req *spacebackupv1.ListBackupsRequest) (*spacebackupv1.ListBackupsResponse, error) {
	f.ListBackupsCalls.record(req)
	return f.ListBackupsCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackups)
}

// GetBackup records the call and dispatches via GetBackupCalls.
func (f *FakeServerlessBackupService) GetBackup(ctx context.Context, req *spacebackupv1.GetBackupRequest) (*spacebackupv1.GetBackupResponse, error) {
	f.GetBackupCalls.record(req)
	return f.GetBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackup)
}

// CreateBackup records the call and dispatches via CreateBackupCalls.
func (f *FakeServerlessBackupService) CreateBackup(ctx context.Context, req *spacebackupv1.CreateBackupRequest) (*spacebackupv1.CreateBackupResponse, error) {
	f.CreateBackupCalls.record(req)
	return f.CreateBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackup)
}

// DeleteBackup records the call and dispatches via DeleteBackupCalls.
func (f *FakeServerlessBackupService) DeleteBackup(ctx context.Context, req *spacebackupv1.DeleteBackupRequest) (*spacebackupv1.DeleteBackupResponse, error) {
	f.DeleteBackupCalls.record(req)
	return f.DeleteBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackup)
}

// ListBackupRestores records the call and dispatches via ListBackupRestoresCalls.
func (f *FakeServerlessBackupService) ListBackupRestores(ctx context.Context, req *spacebackupv1.ListBackupRestoresRequest) (*spacebackupv1.ListBackupRestoresResponse, error) {
	f.ListBackupRestoresCalls.record(req)
	return f.ListBackupRestoresCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupRestores)
}

// RestoreBackup records the call and dispatches via RestoreBackupCalls.
func (f *FakeServerlessBackupService) RestoreBackup(ctx context.Context, req *spacebackupv1.RestoreBackupRequest) (*spacebackupv1.RestoreBackupResponse, error) {
	f.RestoreBackupCalls.record(req)
	return f.RestoreBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.RestoreBackup)
}

// ListBackupSchedules records the call and dispatches via ListBackupSchedulesCalls.
func (f *FakeServerlessBackupService) ListBackupSchedules(ctx context.Context, req *spacebackupv1.ListBackupSchedulesRequest) (*spacebackupv1.ListBackupSchedulesResponse, error) {
	f.ListBackupSchedulesCalls.record(req)
	return f.ListBackupSchedulesCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupSchedules)
}

// GetBackupSchedule records the call and dispatches via GetBackupScheduleCalls.
func (f *FakeServerlessBackupService) GetBackupSchedule(ctx context.Context, req *spacebackupv1.GetBackupScheduleRequest) (*spacebackupv1.GetBackupScheduleResponse, error) {
	f.GetBackupScheduleCalls.record(req)
	return f.GetBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackupSchedule)
}

// CreateBackupSchedule records the call and dispatches via CreateBackupScheduleCalls.
func (f *FakeServerlessBackupService) CreateBackupSchedule(ctx context.Context, req *spacebackupv1.CreateBackupScheduleRequest) (*spacebackupv1.CreateBackupScheduleResponse, error) {
	f.CreateBackupScheduleCalls.record(req)
	return f.CreateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackupSchedule)
}

// UpdateBackupSchedule records the call and dispatches via UpdateBackupScheduleCalls.
func (f *FakeServerlessBackupService) UpdateBackupSchedule(ctx context.Context, req *spacebackupv1.UpdateBackupScheduleRequest) (*spacebackupv1.UpdateBackupScheduleResponse, error) {
	f.UpdateBackupScheduleCalls.record(req)
	return f.UpdateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.UpdateBackupSchedule)
}

// DeleteBackupSchedule records the call and dispatches via DeleteBackupScheduleCalls.
func (f *FakeServerlessBackupService) DeleteBackupSchedule(ctx context.Context, req *spacebackupv1.DeleteBackupScheduleRequest) (*spacebackupv1.DeleteBackupScheduleResponse, error) {
	f.DeleteBackupScheduleCalls.record(req)
	return f.DeleteBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackupSchedule)
}
