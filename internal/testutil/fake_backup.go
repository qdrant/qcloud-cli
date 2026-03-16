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
}

// ListBackups delegates to ListBackupsFunc if set.
func (f *FakeBackupService) ListBackups(ctx context.Context, req *backupv1.ListBackupsRequest) (*backupv1.ListBackupsResponse, error) {
	if f.ListBackupsFunc != nil {
		return f.ListBackupsFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.ListBackups(ctx, req)
}

// GetBackup delegates to GetBackupFunc if set.
func (f *FakeBackupService) GetBackup(ctx context.Context, req *backupv1.GetBackupRequest) (*backupv1.GetBackupResponse, error) {
	if f.GetBackupFunc != nil {
		return f.GetBackupFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.GetBackup(ctx, req)
}

// CreateBackup delegates to CreateBackupFunc if set.
func (f *FakeBackupService) CreateBackup(ctx context.Context, req *backupv1.CreateBackupRequest) (*backupv1.CreateBackupResponse, error) {
	if f.CreateBackupFunc != nil {
		return f.CreateBackupFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.CreateBackup(ctx, req)
}

// DeleteBackup delegates to DeleteBackupFunc if set.
func (f *FakeBackupService) DeleteBackup(ctx context.Context, req *backupv1.DeleteBackupRequest) (*backupv1.DeleteBackupResponse, error) {
	if f.DeleteBackupFunc != nil {
		return f.DeleteBackupFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.DeleteBackup(ctx, req)
}

// ListBackupRestores delegates to ListBackupRestoresFunc if set.
func (f *FakeBackupService) ListBackupRestores(ctx context.Context, req *backupv1.ListBackupRestoresRequest) (*backupv1.ListBackupRestoresResponse, error) {
	if f.ListBackupRestoresFunc != nil {
		return f.ListBackupRestoresFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.ListBackupRestores(ctx, req)
}

// RestoreBackup delegates to RestoreBackupFunc if set.
func (f *FakeBackupService) RestoreBackup(ctx context.Context, req *backupv1.RestoreBackupRequest) (*backupv1.RestoreBackupResponse, error) {
	if f.RestoreBackupFunc != nil {
		return f.RestoreBackupFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.RestoreBackup(ctx, req)
}

// ListBackupSchedules delegates to ListBackupSchedulesFunc if set.
func (f *FakeBackupService) ListBackupSchedules(ctx context.Context, req *backupv1.ListBackupSchedulesRequest) (*backupv1.ListBackupSchedulesResponse, error) {
	if f.ListBackupSchedulesFunc != nil {
		return f.ListBackupSchedulesFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.ListBackupSchedules(ctx, req)
}

// GetBackupSchedule delegates to GetBackupScheduleFunc if set.
func (f *FakeBackupService) GetBackupSchedule(ctx context.Context, req *backupv1.GetBackupScheduleRequest) (*backupv1.GetBackupScheduleResponse, error) {
	if f.GetBackupScheduleFunc != nil {
		return f.GetBackupScheduleFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.GetBackupSchedule(ctx, req)
}

// CreateBackupSchedule delegates to CreateBackupScheduleFunc if set.
func (f *FakeBackupService) CreateBackupSchedule(ctx context.Context, req *backupv1.CreateBackupScheduleRequest) (*backupv1.CreateBackupScheduleResponse, error) {
	if f.CreateBackupScheduleFunc != nil {
		return f.CreateBackupScheduleFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.CreateBackupSchedule(ctx, req)
}

// UpdateBackupSchedule delegates to UpdateBackupScheduleFunc if set.
func (f *FakeBackupService) UpdateBackupSchedule(ctx context.Context, req *backupv1.UpdateBackupScheduleRequest) (*backupv1.UpdateBackupScheduleResponse, error) {
	if f.UpdateBackupScheduleFunc != nil {
		return f.UpdateBackupScheduleFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.UpdateBackupSchedule(ctx, req)
}

// DeleteBackupSchedule delegates to DeleteBackupScheduleFunc if set.
func (f *FakeBackupService) DeleteBackupSchedule(ctx context.Context, req *backupv1.DeleteBackupScheduleRequest) (*backupv1.DeleteBackupScheduleResponse, error) {
	if f.DeleteBackupScheduleFunc != nil {
		return f.DeleteBackupScheduleFunc(ctx, req)
	}
	return f.UnimplementedBackupServiceServer.DeleteBackupSchedule(ctx, req)
}
