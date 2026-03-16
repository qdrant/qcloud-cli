package testutil

import (
	"context"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
)

// FakeBackupService is a test fake that implements BackupServiceServer.
// Use the *Calls fields to configure responses and inspect captured requests.
type FakeBackupService struct {
	backupv1.UnimplementedBackupServiceServer

	ListBackupsCalls          MethodSpy[*backupv1.ListBackupsRequest, *backupv1.ListBackupsResponse]
	GetBackupCalls            MethodSpy[*backupv1.GetBackupRequest, *backupv1.GetBackupResponse]
	CreateBackupCalls         MethodSpy[*backupv1.CreateBackupRequest, *backupv1.CreateBackupResponse]
	DeleteBackupCalls         MethodSpy[*backupv1.DeleteBackupRequest, *backupv1.DeleteBackupResponse]
	ListBackupRestoresCalls   MethodSpy[*backupv1.ListBackupRestoresRequest, *backupv1.ListBackupRestoresResponse]
	RestoreBackupCalls        MethodSpy[*backupv1.RestoreBackupRequest, *backupv1.RestoreBackupResponse]
	ListBackupSchedulesCalls  MethodSpy[*backupv1.ListBackupSchedulesRequest, *backupv1.ListBackupSchedulesResponse]
	GetBackupScheduleCalls    MethodSpy[*backupv1.GetBackupScheduleRequest, *backupv1.GetBackupScheduleResponse]
	CreateBackupScheduleCalls MethodSpy[*backupv1.CreateBackupScheduleRequest, *backupv1.CreateBackupScheduleResponse]
	UpdateBackupScheduleCalls MethodSpy[*backupv1.UpdateBackupScheduleRequest, *backupv1.UpdateBackupScheduleResponse]
	DeleteBackupScheduleCalls MethodSpy[*backupv1.DeleteBackupScheduleRequest, *backupv1.DeleteBackupScheduleResponse]
}

// ListBackups records the call and dispatches via ListBackupsCalls.
func (f *FakeBackupService) ListBackups(ctx context.Context, req *backupv1.ListBackupsRequest) (*backupv1.ListBackupsResponse, error) {
	f.ListBackupsCalls.record(req)
	return f.ListBackupsCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackups)
}

// GetBackup records the call and dispatches via GetBackupCalls.
func (f *FakeBackupService) GetBackup(ctx context.Context, req *backupv1.GetBackupRequest) (*backupv1.GetBackupResponse, error) {
	f.GetBackupCalls.record(req)
	return f.GetBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackup)
}

// CreateBackup records the call and dispatches via CreateBackupCalls.
func (f *FakeBackupService) CreateBackup(ctx context.Context, req *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
	f.CreateBackupCalls.record(req)
	return f.CreateBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackup)
}

// DeleteBackup records the call and dispatches via DeleteBackupCalls.
func (f *FakeBackupService) DeleteBackup(ctx context.Context, req *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error) {
	f.DeleteBackupCalls.record(req)
	return f.DeleteBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackup)
}

// ListBackupRestores records the call and dispatches via ListBackupRestoresCalls.
func (f *FakeBackupService) ListBackupRestores(ctx context.Context, req *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error) {
	f.ListBackupRestoresCalls.record(req)
	return f.ListBackupRestoresCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupRestores)
}

// RestoreBackup records the call and dispatches via RestoreBackupCalls.
func (f *FakeBackupService) RestoreBackup(ctx context.Context, req *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
	f.RestoreBackupCalls.record(req)
	return f.RestoreBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.RestoreBackup)
}

// ListBackupSchedules records the call and dispatches via ListBackupSchedulesCalls.
func (f *FakeBackupService) ListBackupSchedules(ctx context.Context, req *backupv1.ListBackupSchedulesRequest) (*backupv1.ListBackupSchedulesResponse, error) {
	f.ListBackupSchedulesCalls.record(req)
	return f.ListBackupSchedulesCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupSchedules)
}

// GetBackupSchedule records the call and dispatches via GetBackupScheduleCalls.
func (f *FakeBackupService) GetBackupSchedule(ctx context.Context, req *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
	f.GetBackupScheduleCalls.record(req)
	return f.GetBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackupSchedule)
}

// CreateBackupSchedule records the call and dispatches via CreateBackupScheduleCalls.
func (f *FakeBackupService) CreateBackupSchedule(ctx context.Context, req *backupv1.CreateBackupScheduleRequest) (*backupv1.CreateBackupScheduleResponse, error) {
	f.CreateBackupScheduleCalls.record(req)
	return f.CreateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackupSchedule)
}

// UpdateBackupSchedule records the call and dispatches via UpdateBackupScheduleCalls.
func (f *FakeBackupService) UpdateBackupSchedule(ctx context.Context, req *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error) {
	f.UpdateBackupScheduleCalls.record(req)
	return f.UpdateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.UpdateBackupSchedule)
}

// DeleteBackupSchedule records the call and dispatches via DeleteBackupScheduleCalls.
func (f *FakeBackupService) DeleteBackupSchedule(ctx context.Context, req *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
	f.DeleteBackupScheduleCalls.record(req)
	return f.DeleteBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackupSchedule)
}
