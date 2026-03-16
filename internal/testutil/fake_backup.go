package testutil

import (
	"context"

	backupv1 "github.com/qdrant/qdrant-cloud-public-api/gen/go/qdrant/cloud/cluster/backup/v1"
)

// FakeBackupService is a test fake that implements BackupServiceServer.
// Set the function fields to control responses per test.
type FakeBackupService struct {
	backupv1.UnimplementedBackupServiceServer

	ListBackupsFunc          func(context.Context, *backupv1.ListBackupsRequest) (*backupv1.ListBackupsResponse, error)
	GetBackupFunc            func(context.Context, *backupv1.GetBackupRequest) (*backupv1.GetBackupResponse, error)
	CreateBackupFunc         func(context.Context, *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error)
	DeleteBackupFunc         func(context.Context, *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error)
	ListBackupRestoresFunc   func(context.Context, *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error)
	RestoreBackupFunc        func(context.Context, *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error)
	ListBackupSchedulesFunc  func(context.Context, *backupv1.ListBackupSchedulesRequest) (*backupv1.ListBackupSchedulesResponse, error)
	GetBackupScheduleFunc    func(context.Context, *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error)
	CreateBackupScheduleFunc func(context.Context, *backupv1.CreateBackupScheduleRequest) (*backupv1.CreateBackupScheduleResponse, error)
	UpdateBackupScheduleFunc func(context.Context, *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error)
	DeleteBackupScheduleFunc func(context.Context, *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error)

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

// ListBackups delegates to ListBackupsFunc if set, otherwise dispatches via ListBackupsCalls.
func (f *FakeBackupService) ListBackups(ctx context.Context, req *backupv1.ListBackupsRequest) (*backupv1.ListBackupsResponse, error) {
	f.ListBackupsCalls.record(req)
	if f.ListBackupsFunc != nil {
		return f.ListBackupsFunc(ctx, req)
	}
	return f.ListBackupsCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackups)
}

// GetBackup delegates to GetBackupFunc if set, otherwise dispatches via GetBackupCalls.
func (f *FakeBackupService) GetBackup(ctx context.Context, req *backupv1.GetBackupRequest) (*backupv1.GetBackupResponse, error) {
	f.GetBackupCalls.record(req)
	if f.GetBackupFunc != nil {
		return f.GetBackupFunc(ctx, req)
	}
	return f.GetBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackup)
}

// CreateBackup delegates to CreateBackupFunc if set, otherwise dispatches via CreateBackupCalls.
func (f *FakeBackupService) CreateBackup(ctx context.Context, req *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
	f.CreateBackupCalls.record(req)
	if f.CreateBackupFunc != nil {
		return f.CreateBackupFunc(ctx, req)
	}
	return f.CreateBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackup)
}

// DeleteBackup delegates to DeleteBackupFunc if set, otherwise dispatches via DeleteBackupCalls.
func (f *FakeBackupService) DeleteBackup(ctx context.Context, req *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error) {
	f.DeleteBackupCalls.record(req)
	if f.DeleteBackupFunc != nil {
		return f.DeleteBackupFunc(ctx, req)
	}
	return f.DeleteBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackup)
}

// ListBackupRestores delegates to ListBackupRestoresFunc if set, otherwise dispatches via ListBackupRestoresCalls.
func (f *FakeBackupService) ListBackupRestores(ctx context.Context, req *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error) {
	f.ListBackupRestoresCalls.record(req)
	if f.ListBackupRestoresFunc != nil {
		return f.ListBackupRestoresFunc(ctx, req)
	}
	return f.ListBackupRestoresCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupRestores)
}

// RestoreBackup delegates to RestoreBackupFunc if set, otherwise dispatches via RestoreBackupCalls.
func (f *FakeBackupService) RestoreBackup(ctx context.Context, req *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
	f.RestoreBackupCalls.record(req)
	if f.RestoreBackupFunc != nil {
		return f.RestoreBackupFunc(ctx, req)
	}
	return f.RestoreBackupCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.RestoreBackup)
}

// ListBackupSchedules delegates to ListBackupSchedulesFunc if set, otherwise dispatches via ListBackupSchedulesCalls.
func (f *FakeBackupService) ListBackupSchedules(ctx context.Context, req *backupv1.ListBackupSchedulesRequest) (*backupv1.ListBackupSchedulesResponse, error) {
	f.ListBackupSchedulesCalls.record(req)
	if f.ListBackupSchedulesFunc != nil {
		return f.ListBackupSchedulesFunc(ctx, req)
	}
	return f.ListBackupSchedulesCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.ListBackupSchedules)
}

// GetBackupSchedule delegates to GetBackupScheduleFunc if set, otherwise dispatches via GetBackupScheduleCalls.
func (f *FakeBackupService) GetBackupSchedule(ctx context.Context, req *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
	f.GetBackupScheduleCalls.record(req)
	if f.GetBackupScheduleFunc != nil {
		return f.GetBackupScheduleFunc(ctx, req)
	}
	return f.GetBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.GetBackupSchedule)
}

// CreateBackupSchedule delegates to CreateBackupScheduleFunc if set, otherwise dispatches via CreateBackupScheduleCalls.
func (f *FakeBackupService) CreateBackupSchedule(ctx context.Context, req *backupv1.CreateBackupScheduleRequest) (*backupv1.CreateBackupScheduleResponse, error) {
	f.CreateBackupScheduleCalls.record(req)
	if f.CreateBackupScheduleFunc != nil {
		return f.CreateBackupScheduleFunc(ctx, req)
	}
	return f.CreateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.CreateBackupSchedule)
}

// UpdateBackupSchedule delegates to UpdateBackupScheduleFunc if set, otherwise dispatches via UpdateBackupScheduleCalls.
func (f *FakeBackupService) UpdateBackupSchedule(ctx context.Context, req *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error) {
	f.UpdateBackupScheduleCalls.record(req)
	if f.UpdateBackupScheduleFunc != nil {
		return f.UpdateBackupScheduleFunc(ctx, req)
	}
	return f.UpdateBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.UpdateBackupSchedule)
}

// DeleteBackupSchedule delegates to DeleteBackupScheduleFunc if set, otherwise dispatches via DeleteBackupScheduleCalls.
func (f *FakeBackupService) DeleteBackupSchedule(ctx context.Context, req *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
	f.DeleteBackupScheduleCalls.record(req)
	if f.DeleteBackupScheduleFunc != nil {
		return f.DeleteBackupScheduleFunc(ctx, req)
	}
	return f.DeleteBackupScheduleCalls.dispatch(ctx, req, f.UnimplementedBackupServiceServer.DeleteBackupSchedule)
}
