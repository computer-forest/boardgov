import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatTabsModule } from '@angular/material/tabs';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { DocumentService } from '@core/services/api.service';
import { NotificationService } from '@core/services/notification.service';
import { Document, DocumentVersion, DocumentReviewer, DocumentApprover } from '@core/models';

@Component({
  selector: 'bg-document-detail',
  standalone: true,
  imports: [
    CommonModule, RouterLink, ReactiveFormsModule,
    MatCardModule, MatButtonModule, MatIconModule, MatChipsModule,
    MatTabsModule, MatFormFieldModule, MatInputModule, MatSelectModule,
    MatProgressSpinnerModule, MatDividerModule, MatTooltipModule
  ],
  template: `
    <div class="page-container" *ngIf="doc; else loadingTpl">

      <!-- ── Header ─────────────────────────────────────────── -->
      <div class="doc-header fade-in">
        <div>
          <button mat-button routerLink="/documents" class="back-btn">
            <mat-icon>arrow_back</mat-icon> Documents
          </button>
          <div class="doc-title-row">
            <mat-icon *ngIf="doc.locked_at" class="lock-icon" matTooltip="Document locked — approved">lock</mat-icon>
            <h2 class="section-title">{{ doc.title }}</h2>
          </div>
          <div class="doc-meta-row">
            <span class="status-chip status-{{ doc.status.toLowerCase().replace('_','-') }}">
              {{ doc.status | titlecase }}
            </span>
            <span class="meta-pill">
              <mat-icon>{{ doc.workflow_mode === 'SEQUENTIAL' ? 'linear_scale' : 'call_split' }}</mat-icon>
              {{ doc.workflow_mode }}
            </span>
            <span class="meta-pill" *ngIf="doc.locked_at"><mat-icon>lock</mat-icon> Locked</span>
          </div>
        </div>

        <div class="doc-actions" *ngIf="!doc.locked_at">
          <button mat-stroked-button *ngIf="doc.status === 'DRAFT'"
                  (click)="submitDoc()" [disabled]="acting">
            <mat-icon>send</mat-icon> Submit
          </button>
          <button mat-stroked-button *ngIf="doc.status === 'SUBMITTED'"
                  (click)="startReview()" [disabled]="acting">
            <mat-icon>rate_review</mat-icon> Start Review
          </button>
          <button mat-stroked-button color="warn" *ngIf="doc.status === 'APPROVED'"
                  (click)="archive()" [disabled]="acting">
            <mat-icon>archive</mat-icon> Archive
          </button>
        </div>
      </div>

      <!-- ── Workflow stepper ───────────────────────────────── -->
      <div class="workflow-stepper fade-in fade-in-delay-1">
        <div *ngFor="let step of workflowSteps; let i = index" class="wf-step"
             [class.active]="step.status === doc.status"
             [class.done]="isStepDone(step.status)">
          <div class="wf-dot">
            <mat-icon *ngIf="isStepDone(step.status)">check</mat-icon>
            <span *ngIf="!isStepDone(step.status)">{{ i + 1 }}</span>
          </div>
          <span class="wf-label">{{ step.label }}</span>
          <div *ngIf="i < workflowSteps.length - 1" class="wf-line"
               [class.done]="isStepDone(step.status)"></div>
        </div>
      </div>

      <!-- ── Tabs ───────────────────────────────────────────── -->
      <mat-tab-group class="fade-in fade-in-delay-2" animationDuration="200ms">

        <!-- Overview -->
        <mat-tab label="Overview">
          <div class="tab-content">
            <div class="overview-grid">
              <mat-card>
                <mat-card-header><mat-card-title>Description</mat-card-title></mat-card-header>
                <mat-card-content>
                  <p class="desc-text">{{ doc.description || 'No description provided.' }}</p>
                  <mat-divider class="divider"></mat-divider>
                  <div class="meta-grid">
                    <div><span class="meta-key">Created</span><span class="meta-val">{{ doc.created_at | date:'MMM d, y' }}</span></div>
                    <div><span class="meta-key">Updated</span><span class="meta-val">{{ doc.updated_at | date:'MMM d, y' }}</span></div>
                    <div><span class="meta-key">Workflow</span><span class="meta-val">{{ doc.workflow_mode }}</span></div>
                  </div>
                </mat-card-content>
              </mat-card>

              <mat-card *ngIf="doc.status === 'UNDER_REVIEW'" class="approval-card">
                <mat-card-header>
                  <mat-icon mat-card-avatar class="action-icon">how_to_vote</mat-icon>
                  <mat-card-title>Process Approval</mat-card-title>
                  <mat-card-subtitle>Submit your decision on this document</mat-card-subtitle>
                </mat-card-header>
                <mat-card-content>
                  <form [formGroup]="approvalForm" (ngSubmit)="processApproval()">
                    <mat-form-field appearance="outline" class="full-width">
                      <mat-label>Decision</mat-label>
                      <mat-select formControlName="action">
                        <mat-option value="approved">&#10003; Approve</mat-option>
                        <mat-option value="rejected">&#10007; Reject</mat-option>
                      </mat-select>
                    </mat-form-field>
                    <mat-form-field appearance="outline" class="full-width">
                      <mat-label>Comments</mat-label>
                      <textarea matInput formControlName="comments" rows="3"
                                placeholder="Optional comments..."></textarea>
                    </mat-form-field>
                    <button mat-raised-button color="primary" type="submit"
                            [disabled]="approvalForm.invalid || acting" class="full-width">
                      <mat-spinner *ngIf="acting" diameter="18"></mat-spinner>
                      <span *ngIf="!acting">Submit Decision</span>
                    </button>
                  </form>
                </mat-card-content>
              </mat-card>
            </div>
          </div>
        </mat-tab>

        <!-- Reviewers -->
        <mat-tab label="Reviewers">
          <div class="tab-content">

            <!-- Document file for reviewer -->
            <ng-container *ngIf="latestVersion; else noFileTpl">
              <mat-card class="file-panel">
                <mat-card-content class="file-panel-body">
                  <div class="file-panel-left">
                    <mat-icon class="fp-icon" [ngClass]="versionIconClass(latestVersion)">
                      {{ versionIcon(latestVersion) }}
                    </mat-icon>
                    <div>
                      <div class="fp-label">Document to Review</div>
                      <div class="fp-name">{{ latestVersion.file_name }}</div>
                      <div class="fp-meta">
                        {{ formatFileSize(latestVersion.file_size) }}
                        &nbsp;&middot;&nbsp; v{{ latestVersion.version_number }}
                        &nbsp;&middot;&nbsp; {{ latestVersion.created_at | date:'MMM d, y' }}
                      </div>
                    </div>
                  </div>
                  <a mat-stroked-button [href]="downloadUrl(latestVersion.version_number)"
                     target="_blank">
                    <mat-icon>download</mat-icon> Download
                  </a>
                </mat-card-content>
              </mat-card>
            </ng-container>
            <ng-template #noFileTpl>
              <mat-card class="file-panel file-panel-empty">
                <mat-card-content class="file-panel-body">
                  <mat-icon>upload_file</mat-icon>
                  <span class="fp-empty-msg">No file uploaded yet — go to <strong>Files &amp; Versions</strong> to attach one</span>
                </mat-card-content>
              </mat-card>
            </ng-template>

            <!-- Reviewer list -->
            <mat-card>
              <mat-card-header>
                <mat-card-title>Reviewer Status</mat-card-title>
                <mat-card-subtitle>{{ reviewers.length }} reviewer{{ reviewers.length !== 1 ? 's' : '' }} assigned</mat-card-subtitle>
              </mat-card-header>
              <mat-card-content>
                <div *ngIf="reviewers.length === 0" class="empty-state">
                  <mat-icon>group</mat-icon><span>No reviewers assigned yet</span>
                </div>
                <div *ngFor="let r of reviewers; let i = index" class="person-row">
                  <div class="order-badge">{{ i + 1 }}</div>
                  <div class="person-avatar">{{ r.user_id | slice:0:2 | uppercase }}</div>
                  <div class="person-info">
                    <div class="person-name">Reviewer {{ r.user_id | slice:0:8 }}</div>
                    <div class="person-meta" *ngIf="r.reviewed_at">
                      Reviewed: {{ r.reviewed_at | date:'MMM d, y HH:mm' }}
                    </div>
                    <div class="person-comment" *ngIf="r.comments">{{ r.comments }}</div>
                  </div>
                  <span class="status-badge" [ngClass]="r.status">
                    <mat-icon>{{ statusIcon(r.status) }}</mat-icon>
                    {{ r.status | titlecase }}
                  </span>
                </div>
              </mat-card-content>
            </mat-card>

          </div>
        </mat-tab>

        <!-- Files & Versions -->
        <mat-tab label="Files &amp; Versions">
          <div class="tab-content">
            <div class="upload-zone" *ngIf="!doc.locked_at"
                 (dragover)="$event.preventDefault()"
                 (drop)="onDrop($event)"
                 (click)="fileInput.click()">
              <mat-icon>cloud_upload</mat-icon>
              <div>
                <p class="upload-title">Drop files here or click to upload</p>
                <p class="upload-hint">PDF, ZIP, Word (.doc, .docx) &nbsp;&middot;&nbsp; Max 30 MB</p>
              </div>
              <input #fileInput type="file" accept=".pdf,.zip,.doc,.docx"
                     style="display:none" (change)="onFileSelect($event)" />
            </div>

            <div *ngIf="uploading" class="upload-progress">
              <mat-progress-spinner mode="indeterminate" diameter="24"></mat-progress-spinner>
              <span>Uploading&hellip;</span>
            </div>

            <mat-card>
              <mat-card-header><mat-card-title>Version History</mat-card-title></mat-card-header>
              <mat-card-content>
                <div *ngIf="versions.length === 0" class="empty-state">
                  <mat-icon>folder_open</mat-icon><span>No files uploaded yet</span>
                </div>
                <div *ngFor="let v of versions" class="version-row">
                  <div class="version-badge">v{{ v.version_number }}</div>
                  <mat-icon class="ver-file-icon" [ngClass]="versionIconClass(v)">{{ versionIcon(v) }}</mat-icon>
                  <div class="version-info">
                    <div class="version-name">{{ v.file_name }}</div>
                    <div class="version-meta">
                      {{ formatFileSize(v.file_size) }} &middot; {{ v.mime_type }} &middot; {{ v.created_at | date:'MMM d, y HH:mm' }}
                    </div>
                    <div class="version-hash mono">SHA256: {{ v.sha256_hash | slice:0:16 }}&hellip;</div>
                  </div>
                  <a mat-icon-button [href]="downloadUrl(v.version_number)"
                     target="_blank" matTooltip="Download">
                    <mat-icon>download</mat-icon>
                  </a>
                </div>
              </mat-card-content>
            </mat-card>
          </div>
        </mat-tab>

        <!-- Approvals -->
        <mat-tab label="Approvals">
          <div class="tab-content">

            <!-- Document file for approver -->
            <ng-container *ngIf="latestVersion; else noFileTpl2">
              <mat-card class="file-panel">
                <mat-card-content class="file-panel-body">
                  <div class="file-panel-left">
                    <mat-icon class="fp-icon" [ngClass]="versionIconClass(latestVersion)">
                      {{ versionIcon(latestVersion) }}
                    </mat-icon>
                    <div>
                      <div class="fp-label">Document to Approve</div>
                      <div class="fp-name">{{ latestVersion.file_name }}</div>
                      <div class="fp-meta">
                        {{ formatFileSize(latestVersion.file_size) }}
                        &nbsp;&middot;&nbsp; v{{ latestVersion.version_number }}
                        &nbsp;&middot;&nbsp; {{ latestVersion.created_at | date:'MMM d, y' }}
                      </div>
                    </div>
                  </div>
                  <a mat-stroked-button [href]="downloadUrl(latestVersion.version_number)"
                     target="_blank">
                    <mat-icon>download</mat-icon> Download
                  </a>
                </mat-card-content>
              </mat-card>
            </ng-container>
            <ng-template #noFileTpl2>
              <mat-card class="file-panel file-panel-empty">
                <mat-card-content class="file-panel-body">
                  <mat-icon>upload_file</mat-icon>
                  <span class="fp-empty-msg">No file uploaded yet — go to <strong>Files &amp; Versions</strong> to attach one</span>
                </mat-card-content>
              </mat-card>
            </ng-template>

            <!-- Approver list -->
            <mat-card>
              <mat-card-header>
                <mat-card-title>Approver Status</mat-card-title>
                <mat-card-subtitle>{{ approvers.length }} approver{{ approvers.length !== 1 ? 's' : '' }} assigned</mat-card-subtitle>
              </mat-card-header>
              <mat-card-content>
                <div *ngIf="approvers.length === 0" class="empty-state">
                  <mat-icon>how_to_vote</mat-icon><span>No approvers assigned</span>
                </div>
                <div *ngFor="let a of approvers; let i = index" class="person-row">
                  <div class="order-badge">{{ i + 1 }}</div>
                  <div class="person-avatar">{{ a.user_id | slice:0:2 | uppercase }}</div>
                  <div class="person-info">
                    <div class="person-name">Approver {{ a.user_id | slice:0:8 }}</div>
                    <div class="person-meta" *ngIf="a.approved_at">
                      Actioned: {{ a.approved_at | date:'MMM d, y HH:mm' }}
                    </div>
                    <div class="person-comment" *ngIf="a.comments">{{ a.comments }}</div>
                  </div>
                  <span class="status-badge" [ngClass]="a.status">
                    <mat-icon>{{ statusIcon(a.status) }}</mat-icon>
                    {{ a.status | titlecase }}
                  </span>
                </div>
              </mat-card-content>
            </mat-card>

          </div>
        </mat-tab>

      </mat-tab-group>
    </div>

    <ng-template #loadingTpl>
      <div class="loading-center"><mat-spinner></mat-spinner></div>
    </ng-template>
  `,
  styles: [`
    /* Header */
    .doc-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:20px; flex-wrap:wrap; gap:16px; }
    .back-btn { color:var(--text-muted); margin-bottom:6px; padding-left:0; }
    .doc-title-row { display:flex; align-items:center; gap:10px; margin-bottom:8px; }
    .lock-icon { color:var(--gold); font-size:22px!important; }
    .doc-meta-row { display:flex; align-items:center; gap:10px; flex-wrap:wrap; }
    .doc-actions { display:flex; gap:10px; flex-shrink:0; }
    .status-chip { font-size:.75rem; padding:3px 10px; border-radius:100px; font-weight:600; }
    .meta-pill { display:flex; align-items:center; gap:4px; font-size:.75rem; color:var(--text-muted); background:var(--bg-elevated); padding:3px 10px; border-radius:100px; mat-icon{font-size:14px!important;} }

    /* Stepper */
    .workflow-stepper { display:flex; align-items:center; margin-bottom:24px; background:var(--bg-card); border:1px solid var(--border); border-radius:var(--radius-lg); padding:20px 24px; flex-wrap:wrap; gap:4px; }
    .wf-step { display:flex; align-items:center; gap:8px; }
    .wf-dot { width:28px; height:28px; border-radius:50%; border:2px solid var(--border); display:flex; align-items:center; justify-content:center; font-size:12px; font-weight:600; color:var(--text-muted); transition:all var(--transition); mat-icon{font-size:14px!important;width:14px;height:14px;} }
    .wf-step.done   .wf-dot { border-color:var(--success); background:rgba(34,197,94,.15); color:var(--success); }
    .wf-step.active .wf-dot { border-color:var(--gold); background:var(--gold-dim); color:var(--gold); }
    .wf-label { font-size:.75rem; color:var(--text-muted); white-space:nowrap; }
    .wf-step.active .wf-label { color:var(--gold); font-weight:600; }
    .wf-step.done   .wf-label { color:var(--success); }
    .wf-line { width:40px; height:2px; background:var(--border); margin:0 4px; border-radius:1px; &.done{background:var(--success);} }

    /* Tabs */
    .tab-content { padding:20px 0; display:flex; flex-direction:column; gap:20px; }

    /* Overview */
    .overview-grid { display:grid; grid-template-columns:1fr 1fr; gap:20px; }
    @media(max-width:900px) { .overview-grid { grid-template-columns:1fr; } }
    .desc-text { color:var(--text-secondary); font-size:.875rem; line-height:1.7; }
    .divider { margin:16px 0; }
    .meta-grid { display:grid; grid-template-columns:repeat(3,1fr); gap:16px; margin-top:16px; }
    .meta-key { display:block; font-size:.7rem; text-transform:uppercase; letter-spacing:.06em; color:var(--text-muted); margin-bottom:2px; }
    .meta-val { font-size:.875rem; color:var(--text-primary); font-weight:500; }
    .action-icon { background:var(--gold-dim)!important; color:var(--gold)!important; width:36px!important; height:36px!important; border-radius:10px!important; display:flex!important; align-items:center; justify-content:center; }
    .approval-card { border-color:var(--gold-border)!important; }
    .full-width { width:100%; margin-bottom:12px; display:block; }

    /* File panel (Reviewers + Approvals tabs) */
    .file-panel { border-left:4px solid var(--gold-border)!important; }
    .file-panel-body { display:flex; align-items:center; justify-content:space-between; gap:16px; flex-wrap:wrap; padding:16px!important; }
    .file-panel-left { display:flex; align-items:center; gap:14px; }
    .fp-icon { font-size:40px!important; width:40px; height:40px; flex-shrink:0; }
    .fp-icon.pdf { color:#ef4444; }
    .fp-icon.zip { color:#f59e0b; }
    .fp-icon.doc { color:#3b82f6; }
    .fp-label { font-size:.68rem; font-weight:700; text-transform:uppercase; letter-spacing:.06em; color:var(--text-muted); margin-bottom:3px; }
    .fp-name  { font-size:.95rem; font-weight:600; color:var(--text-primary); }
    .fp-meta  { font-size:.75rem; color:var(--text-muted); margin-top:2px; }
    .file-panel-empty .file-panel-body { justify-content:flex-start; color:var(--text-muted); mat-icon{opacity:.4; font-size:28px!important;width:28px;height:28px;} }
    .fp-empty-msg { font-size:.85rem; }

    /* People rows (Reviewers & Approvals) */
    .person-row { display:flex; align-items:flex-start; gap:12px; padding:12px 0; border-bottom:1px solid var(--border); &:last-child{border:none;} }
    .order-badge { width:22px; height:22px; border-radius:50%; background:var(--bg-elevated); display:flex; align-items:center; justify-content:center; font-size:.7rem; font-weight:700; color:var(--text-muted); flex-shrink:0; margin-top:6px; }
    .person-avatar { width:36px; height:36px; border-radius:50%; background:var(--bg-elevated); display:flex; align-items:center; justify-content:center; font-size:.75rem; font-weight:700; color:var(--text-muted); flex-shrink:0; }
    .person-info { flex:1; }
    .person-name    { font-size:.875rem; font-weight:500; color:var(--text-primary); }
    .person-meta    { font-size:.75rem; color:var(--text-muted); margin-top:2px; }
    .person-comment { font-size:.8rem; color:var(--text-secondary); font-style:italic; margin-top:4px; }
    .status-badge {
      display:flex; align-items:center; gap:4px;
      font-size:.75rem; font-weight:600; flex-shrink:0;
      padding:4px 10px; border-radius:100px;
      mat-icon{ font-size:15px!important; width:15px; height:15px; }
      &.pending  { color:var(--warning); background:rgba(245,158,11,.1); }
      &.approved { color:var(--success); background:rgba(34,197,94,.1); }
      &.rejected { color:var(--danger);  background:rgba(239,68,68,.1); }
    }

    /* Files tab */
    .upload-zone { border:2px dashed var(--border); border-radius:var(--radius-lg); padding:36px; text-align:center; cursor:pointer; transition:all var(--transition); display:flex; flex-direction:column; align-items:center; gap:12px;
      mat-icon{font-size:44px!important;width:44px;height:44px;color:var(--text-muted);opacity:.5;}
      &:hover{border-color:var(--gold-border);background:var(--gold-dim);}
    }
    .upload-title { font-weight:500; color:var(--text-primary); margin-bottom:4px; }
    .upload-hint  { font-size:.8rem; color:var(--text-muted); }
    .upload-progress { display:flex; align-items:center; gap:12px; padding:12px; }
    .version-row { display:flex; align-items:center; gap:14px; padding:12px 0; border-bottom:1px solid var(--border); &:last-child{border:none;} }
    .version-badge { background:var(--gold-dim); color:var(--gold); font-size:.8rem; font-weight:700; padding:4px 10px; border-radius:var(--radius-sm); font-family:'IBM Plex Mono',monospace; flex-shrink:0; }
    .ver-file-icon { font-size:28px!important; width:28px; height:28px; flex-shrink:0; &.pdf{color:#ef4444;} &.zip{color:#f59e0b;} &.doc{color:#3b82f6;} }
    .version-info { flex:1; }
    .version-name { font-size:.875rem; font-weight:500; color:var(--text-primary); }
    .version-meta { font-size:.75rem; color:var(--text-muted); margin-top:2px; }
    .version-hash { font-size:.65rem; color:var(--text-muted); opacity:.6; margin-top:2px; }

    /* Misc */
    .empty-state { display:flex; flex-direction:column; align-items:center; gap:8px; padding:40px; color:var(--text-muted); mat-icon{font-size:36px;width:36px;height:36px;opacity:.3;} }
    .loading-center { display:flex; justify-content:center; padding:80px; }
    .mono { font-family:'IBM Plex Mono',monospace; }
  `]
})
export class DocumentDetailComponent implements OnInit {
  private route  = inject(ActivatedRoute);
  private docSvc = inject(DocumentService);
  private notif  = inject(NotificationService);
  private fb     = inject(FormBuilder);

  doc: Document | null           = null;
  versions:  DocumentVersion[]   = [];
  reviewers: DocumentReviewer[]  = [];
  approvers: DocumentApprover[]  = [];
  acting    = false;
  uploading = false;

  workflowSteps = [
    { label: 'Draft',        status: 'DRAFT' },
    { label: 'Submitted',    status: 'SUBMITTED' },
    { label: 'Under Review', status: 'UNDER_REVIEW' },
    { label: 'Approved',     status: 'APPROVED' },
    { label: 'Archived',     status: 'ARCHIVED' },
  ];

  approvalForm = this.fb.nonNullable.group({
    action:   ['approved', Validators.required],
    comments: ['']
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.load(id);
  }

  load(id: string): void {
    this.docSvc.getById(id).subscribe(d => {
      this.doc = d;
      this.docSvc.listVersions(id).subscribe(v => { this.versions  = v  ?? []; });
      this.docSvc.getReviewers(id).subscribe(r => { this.reviewers = r  ?? []; });
      this.docSvc.getApprovers(id).subscribe(a => { this.approvers = a  ?? []; });
    });
  }

  get docId(): string { return this.doc?.id ?? ''; }

  /** Latest version shown in Reviewers and Approvals file panel. */
  get latestVersion(): DocumentVersion | null {
    return this.versions.length ? this.versions[0] : null;
  }

  isStepDone(status: string): boolean {
    const order = ['DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'APPROVED', 'ARCHIVED'];
    return order.indexOf(this.doc?.status ?? '') > order.indexOf(status);
  }

  statusIcon(status: string): string {
    if (status === 'approved') return 'check_circle';
    if (status === 'rejected') return 'cancel';
    return 'schedule';
  }

  versionIcon(v: DocumentVersion): string {
    const n = v.file_name.toLowerCase();
    if (n.endsWith('.pdf'))                        return 'picture_as_pdf';
    if (n.endsWith('.zip'))                        return 'folder_zip';
    if (n.endsWith('.doc') || n.endsWith('.docx')) return 'description';
    return 'insert_drive_file';
  }

  versionIconClass(v: DocumentVersion): string {
    const n = v.file_name.toLowerCase();
    if (n.endsWith('.pdf')) return 'pdf';
    if (n.endsWith('.zip')) return 'zip';
    return 'doc';
  }

  formatFileSize(bytes: number): string {
    if (bytes < 1024)        return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  downloadUrl(version: number): string {
    return this.docSvc.downloadUrl(this.docId, version);
  }

  submitDoc(): void {
    this.acting = true;
    this.docSvc.submit(this.docId).subscribe({
      next: () => { this.notif.success('Document submitted'); this.load(this.docId); this.acting = false; },
      error: err => { this.acting = false; this.notif.error(err.error?.error ?? 'Failed'); }
    });
  }

  startReview(): void {
    this.acting = true;
    this.docSvc.startReview(this.docId).subscribe({
      next: () => { this.notif.success('Review started'); this.load(this.docId); this.acting = false; },
      error: err => { this.acting = false; this.notif.error(err.error?.error ?? 'Failed'); }
    });
  }

  archive(): void {
    this.acting = true;
    this.docSvc.archive(this.docId).subscribe({
      next: () => { this.notif.success('Archived'); this.load(this.docId); this.acting = false; },
      error: err => { this.acting = false; this.notif.error(err.error?.error ?? 'Failed'); }
    });
  }

  processApproval(): void {
    if (this.approvalForm.invalid) return;
    this.acting = true;
    this.docSvc.processApproval(this.docId, this.approvalForm.getRawValue() as any).subscribe({
      next: () => {
        this.notif.success('Decision recorded');
        this.load(this.docId);
        this.acting = false;
        this.approvalForm.reset({ action: 'approved', comments: '' });
      },
      error: err => { this.acting = false; this.notif.error(err.error?.error ?? 'Failed'); }
    });
  }

  onFileSelect(event: Event): void {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) this.uploadFile(file);
  }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    const file = event.dataTransfer?.files?.[0];
    if (file) this.uploadFile(file);
  }

  uploadFile(file: File): void {
    if (file.size > 30 * 1024 * 1024) { this.notif.error('File exceeds 30 MB limit'); return; }
    this.uploading = true;
    this.docSvc.upload(this.docId, file).subscribe({
      next: v => {
        this.versions = [v, ...this.versions];
        this.notif.success(`v${v.version_number} uploaded`);
        this.uploading = false;
      },
      error: err => { this.uploading = false; this.notif.error(err.error?.error ?? 'Upload failed'); }
    });
  }
}
