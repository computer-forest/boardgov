// ─── Auth ─────────────────────────────────────────────────────
export interface User {
  id: string;
  email: string;
  full_name: string;
  is_active: boolean;
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  user: User;
}

// ─── Meetings ─────────────────────────────────────────────────
export type MeetingType = 'BOARD' | 'COMMITTEE' | 'AGM' | 'SPECIAL';

export interface Meeting {
  id: string;
  title: string;
  meeting_type: MeetingType;
  scheduled_at: string;
  location: string;
  agenda: string;
  created_by: string;
  is_locked: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateMeetingRequest {
  title: string;
  meeting_type: MeetingType;
  scheduled_at: string;
  location: string;
  agenda: string;
}

// ─── Documents ────────────────────────────────────────────────
export type DocumentStatus = 'DRAFT' | 'SUBMITTED' | 'UNDER_REVIEW' | 'APPROVED' | 'REJECTED' | 'ARCHIVED';
export type WorkflowMode = 'SEQUENTIAL' | 'PARALLEL';
export type ApprovalStatus = 'PENDING' | 'APPROVED' | 'REJECTED';

export interface Document {
  id: string;
  meeting_id: string;
  title: string;
  description: string;
  status: DocumentStatus;
  workflow_mode: WorkflowMode;
  submitted_at?: string;
  locked_at?: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface DocumentVersion {
  id: string;
  document_id: string;
  version_number: number;
  file_name: string;
  file_size: number;
  mime_type: string;
  sha256_hash: string;
  uploaded_by: string;
  created_at: string;
}

export type ReviewerStatus = 'pending' | 'approved' | 'rejected';
export type ApproverStatus = 'pending' | 'approved' | 'rejected';

export interface DocumentReviewer {
  id: string;
  document_id: string;
  user_id: string;
  order_index: number;
  status: ReviewerStatus;
  reviewed_at?: string;
  comments?: string;
  created_at: string;
}

export interface DocumentApprover {
  id: string;
  document_id: string;
  user_id: string;
  order_index: number;
  status: ApproverStatus;
  comments?: string;
  approved_at?: string;
  created_at: string;
}

export interface Approval {
  id: string;
  document_id: string;
  approver_id: string;
  status: ApprovalStatus;
  comments: string;
  actioned_at?: string;
  created_at: string;
}

export interface CreateDocumentRequest {
  meeting_id: string;
  title: string;
  description: string;
  workflow_mode: WorkflowMode;
}

export interface ApprovalRequest {
  action: 'approved' | 'rejected';
  comments: string;
}

// ─── Dashboard ────────────────────────────────────────────────
export interface PendingItem {
  document_id: string;
  document_title: string;
  meeting_id: string;
  meeting_title: string;
  status: DocumentStatus;
  assigned_at: string;
  due_date?: string;
}

export interface OverdueDocument {
  document_id: string;
  document_title: string;
  meeting_title: string;
  status: DocumentStatus;
  days_overdue: number;
}

export interface MeetingStatusSummary {
  meeting_id: string;
  meeting_title: string;
  meeting_type: MeetingType;
  total: number;
  draft: number;
  submitted: number;
  under_review: number;
  approved: number;
  rejected: number;
  archived: number;
}

export interface AgingBucket {
  bucket: string;
  count: number;
}

export interface TurnaroundMetrics {
  avg_approval_hours: number;
  p50_approval_hours: number;
  p90_approval_hours: number;
  avg_review_hours: number;
  p50_review_hours: number;
  p90_review_hours: number;
}

export interface RejectionAnalytics {
  total_documents: number;
  total_rejected: number;
  rejection_rate: number;
}

// ─── Audit Log ────────────────────────────────────────────────
export interface AuditLog {
  id: string;
  user_id?: string;
  action: string;
  meeting_id?: string;
  document_id?: string;
  previous_state?: DocumentStatus;
  new_state?: DocumentStatus;
  ip_address: string;
  hash: string;
  previous_hash: string;
  timestamp: string;
}

export interface AuditIntegrity {
  valid: boolean;
  errors: string[];
}

// ─── API Pagination ───────────────────────────────────────────
export interface PagedResponse<T> {
  data: T[];
  total: number;
  offset: number;
  limit: number;
}
