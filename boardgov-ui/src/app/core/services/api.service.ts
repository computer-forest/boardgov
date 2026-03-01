import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '@env/environment';
import {
  Meeting, CreateMeetingRequest, PagedResponse,
  Document, CreateDocumentRequest, DocumentVersion,
  DocumentReviewer, DocumentApprover, Approval, ApprovalRequest,
  PendingItem, OverdueDocument, MeetingStatusSummary,
  AgingBucket, TurnaroundMetrics, RejectionAnalytics,
  AuditLog, AuditIntegrity, User
} from '@core/models';

// ─── Meeting Service ──────────────────────────────────────────
@Injectable({ providedIn: 'root' })
export class MeetingService {
  private base = `${environment.apiUrl}/meetings`;
  constructor(private http: HttpClient) {}

  list(page = 1, pageSize = 20): Observable<PagedResponse<Meeting>> {
    return this.http.get<PagedResponse<Meeting>>(this.base, {
      params: new HttpParams().set('page', page).set('page_size', pageSize)
    });
  }

  getById(id: string): Observable<Meeting> {
    return this.http.get<Meeting>(`${this.base}/${id}`);
  }

  create(req: CreateMeetingRequest): Observable<Meeting> {
    return this.http.post<Meeting>(this.base, req);
  }
}

// ─── Document Service ─────────────────────────────────────────
@Injectable({ providedIn: 'root' })
export class DocumentService {
  private base = `${environment.apiUrl}/documents`;
  constructor(private http: HttpClient) {}

  listByMeeting(meetingId: string): Observable<Document[]> {
    return this.http.get<Document[]>(`${environment.apiUrl}/meetings/${meetingId}/documents`);
  }

  getById(id: string): Observable<Document> {
    return this.http.get<Document>(`${this.base}/${id}`);
  }

  create(req: CreateDocumentRequest): Observable<Document> {
    return this.http.post<Document>(this.base, req);
  }

  upload(docId: string, file: File): Observable<DocumentVersion> {
    const form = new FormData();
    form.append('file', file);
    return this.http.post<DocumentVersion>(`${this.base}/${docId}/upload`, form);
  }

  listVersions(docId: string): Observable<DocumentVersion[]> {
    return this.http.get<DocumentVersion[]>(`${this.base}/${docId}/versions`);
  }

  assignReviewers(docId: string, userIds: string[]): Observable<void> {
    return this.http.put<void>(`${this.base}/${docId}/reviewers`, { user_ids: userIds });
  }

  assignApprovers(docId: string, userIds: string[]): Observable<void> {
    return this.http.put<void>(`${this.base}/${docId}/approvers`, { user_ids: userIds });
  }

  getReviewers(docId: string): Observable<DocumentReviewer[]> {
    return this.http.get<DocumentReviewer[]>(`${this.base}/${docId}/reviewers`);
  }

  getApprovers(docId: string): Observable<DocumentApprover[]> {
    return this.http.get<DocumentApprover[]>(`${this.base}/${docId}/approvers`);
  }

  submit(docId: string): Observable<void> {
    return this.http.post<void>(`${this.base}/${docId}/submit`, {});
  }

  startReview(docId: string): Observable<void> {
    return this.http.post<void>(`${this.base}/${docId}/review/start`, {});
  }

  processApproval(docId: string, req: ApprovalRequest): Observable<void> {
    return this.http.post<void>(`${this.base}/${docId}/approve`, req);
  }

  archive(docId: string): Observable<void> {
    return this.http.post<void>(`${this.base}/${docId}/archive`, {});
  }

  downloadUrl(docId: string, version?: number): string {
    const v = version ? `?version=${version}` : '';
    return `${this.base}/${docId}/download${v}`;
  }
}

// ─── Dashboard Service ────────────────────────────────────────
@Injectable({ providedIn: 'root' })
export class DashboardService {
  private base = `${environment.apiUrl}/dashboard`;
  constructor(private http: HttpClient) {}

  pendingApprovals(offset = 0, limit = 10): Observable<PagedResponse<PendingItem>> {
    return this.http.get<PagedResponse<PendingItem>>(`${this.base}/pending-approvals`, {
      params: new HttpParams().set('offset', offset).set('limit', limit)
    });
  }

  pendingReviews(offset = 0, limit = 10): Observable<PagedResponse<PendingItem>> {
    return this.http.get<PagedResponse<PendingItem>>(`${this.base}/pending-reviews`, {
      params: new HttpParams().set('offset', offset).set('limit', limit)
    });
  }

  overdueDocuments(offset = 0, limit = 10): Observable<PagedResponse<OverdueDocument>> {
    return this.http.get<PagedResponse<OverdueDocument>>(`${this.base}/overdue`, {
      params: new HttpParams().set('offset', offset).set('limit', limit)
    });
  }

  meetingStatusSummary(): Observable<{ data: MeetingStatusSummary[] }> {
    return this.http.get<{ data: MeetingStatusSummary[] }>(`${this.base}/meeting-summary`);
  }

  agingAnalysis(): Observable<{ data: AgingBucket[] }> {
    return this.http.get<{ data: AgingBucket[] }>(`${this.base}/aging`);
  }

  turnaroundMetrics(): Observable<TurnaroundMetrics> {
    return this.http.get<TurnaroundMetrics>(`${this.base}/turnaround`);
  }

  rejectionRatio(): Observable<RejectionAnalytics> {
    return this.http.get<RejectionAnalytics>(`${this.base}/rejection-ratio`);
  }

  activityTimeline(offset = 0, limit = 20): Observable<PagedResponse<AuditLog>> {
    return this.http.get<PagedResponse<AuditLog>>(`${this.base}/timeline`, {
      params: new HttpParams().set('offset', offset).set('limit', limit)
    });
  }

  auditIntegrity(): Observable<AuditIntegrity> {
    return this.http.get<AuditIntegrity>(`${this.base}/audit-integrity`);
  }
}

// ─── Audit Service ────────────────────────────────────────────
@Injectable({ providedIn: 'root' })
export class AuditService {
  private base = `${environment.apiUrl}/dashboard`;
  constructor(private http: HttpClient) {}

  list(offset = 0, limit = 20): Observable<PagedResponse<AuditLog>> {
    return this.http.get<PagedResponse<AuditLog>>(`${this.base}/timeline`, {
      params: new HttpParams().set('offset', offset).set('limit', limit)
    });
  }

  integrity(): Observable<AuditIntegrity> {
    return this.http.get<AuditIntegrity>(`${this.base}/audit-integrity`);
  }
}

// ─── User Service ─────────────────────────────────────────────
@Injectable({ providedIn: 'root' })
export class UserService {
  constructor(private http: HttpClient) {}

  list(): Observable<PagedResponse<User>> {
    return this.http.get<PagedResponse<User>>(`${environment.apiUrl}/users`);
  }
}
