import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTableModule } from '@angular/material/table';
import { MatChipsModule } from '@angular/material/chips';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatBadgeModule } from '@angular/material/badge';
import { MatDividerModule } from '@angular/material/divider';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';
import { DashboardService } from '@core/services/api.service';
import { AuthService } from '@core/auth/auth.service';
import {
  PendingItem, OverdueDocument, MeetingStatusSummary,
  AgingBucket, TurnaroundMetrics, RejectionAnalytics, AuditLog
} from '@core/models';
import { ReplacePipe } from '@shared/pipes/replace.pipe';

@Component({
  selector: 'bg-dashboard',
  standalone: true,
  imports: [
    CommonModule, RouterLink,
    MatCardModule, MatIconModule, MatButtonModule,
    MatTableModule, MatChipsModule, MatProgressBarModule,
    MatBadgeModule, MatDividerModule, ReplacePipe
  ],
  template: `
    <div class="page-container">
      <!-- Page header -->
      <div class="page-header fade-in">
        <div>
          <h2 class="section-title">Governance Dashboard</h2>
          <p class="section-subtitle">{{ today | date:'EEEE, MMMM d, y' }} — Board &amp; Committee Overview</p>
        </div>
        <div class="header-actions">
          <button mat-stroked-button color="primary" routerLink="/meetings/new">
            <mat-icon>add</mat-icon> New Meeting
          </button>
          <button mat-raised-button color="primary" routerLink="/documents">
            <mat-icon>description</mat-icon> Documents
          </button>
        </div>
      </div>

      <!-- KPI stat cards -->
      <div class="stat-grid fade-in fade-in-delay-1">
        <div class="stat-card">
          <div class="stat-icon pending"><mat-icon>pending_actions</mat-icon></div>
          <div class="stat-value">{{ pendingApprovals.length }}</div>
          <div class="stat-label">Pending Approvals</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon review"><mat-icon>rate_review</mat-icon></div>
          <div class="stat-value">{{ pendingReviews.length }}</div>
          <div class="stat-label">Pending Reviews</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon overdue"><mat-icon>warning_amber</mat-icon></div>
          <div class="stat-value">{{ overdueDocuments.length }}</div>
          <div class="stat-label">Overdue Documents</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon metrics"><mat-icon>speed</mat-icon></div>
          <div class="stat-value">{{ metrics?.avg_approval_hours | number:'1.0-1' }}h</div>
          <div class="stat-label">Avg Approval Time</div>
        </div>
        <div class="stat-card">
          <div class="stat-icon rejection"><mat-icon>cancel</mat-icon></div>
          <div class="stat-value">{{ rejection?.rejection_rate | number:'1.0-1' }}%</div>
          <div class="stat-label">Rejection Rate</div>
        </div>
      </div>

      <!-- Two-column layout -->
      <div class="dashboard-grid fade-in fade-in-delay-2">

        <!-- Pending Approvals -->
        <mat-card class="dash-card">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon orange">pending_actions</mat-icon>
            <mat-card-title>Pending Approvals</mat-card-title>
            <mat-card-subtitle>Awaiting your action</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div *ngIf="pendingApprovals.length === 0" class="empty-state">
              <mat-icon>check_circle_outline</mat-icon>
              <span>All caught up!</span>
            </div>
            <div *ngFor="let item of pendingApprovals" class="pending-item" [routerLink]="['/documents', item.document_id]">
              <div class="pending-title">{{ item.document_title }}</div>
              <div class="pending-meta">
                <mat-icon class="meta-icon">event</mat-icon>{{ item.meeting_title }}
                <span class="pending-date">{{ item.assigned_at | date:'MMM d' }}</span>
              </div>
              <span class="status-chip status-{{ item.status.toLowerCase() }}">{{ item.status | titlecase }}</span>
            </div>
          </mat-card-content>
          <mat-card-actions>
            <a routerLink="/documents" mat-button color="primary">View All →</a>
          </mat-card-actions>
        </mat-card>

        <!-- Pending Reviews -->
        <mat-card class="dash-card">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon blue">rate_review</mat-icon>
            <mat-card-title>Pending Reviews</mat-card-title>
            <mat-card-subtitle>Documents to review</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div *ngIf="pendingReviews.length === 0" class="empty-state">
              <mat-icon>check_circle_outline</mat-icon>
              <span>No reviews pending</span>
            </div>
            <div *ngFor="let item of pendingReviews" class="pending-item" [routerLink]="['/documents', item.document_id]">
              <div class="pending-title">{{ item.document_title }}</div>
              <div class="pending-meta">
                <mat-icon class="meta-icon">event</mat-icon>{{ item.meeting_title }}
                <span class="pending-date">{{ item.assigned_at | date:'MMM d' }}</span>
              </div>
            </div>
          </mat-card-content>
          <mat-card-actions>
            <a routerLink="/documents" mat-button color="primary">View All →</a>
          </mat-card-actions>
        </mat-card>

        <!-- Meeting Status Summary -->
        <mat-card class="dash-card span-2">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon gold">summarize</mat-icon>
            <mat-card-title>Meeting Status Summary</mat-card-title>
            <mat-card-subtitle>Document progress per meeting</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div *ngFor="let m of meetingSummaries" class="meeting-row">
              <div class="meeting-info">
                <span class="meeting-type-badge">{{ m.meeting_type }}</span>
                <span class="meeting-name">{{ m.meeting_title }}</span>
              </div>
              <div class="meeting-bars">
                <div class="bar-item" *ngIf="m.approved">
                  <span class="bar-label success">Approved</span>
                  <mat-progress-bar mode="determinate" [value]="pct(m.approved, m.total)" color="primary" class="bar-success"></mat-progress-bar>
                  <span class="bar-count">{{ m.approved }}</span>
                </div>
                <div class="bar-item" *ngIf="m.under_review">
                  <span class="bar-label warning">Review</span>
                  <mat-progress-bar mode="determinate" [value]="pct(m.under_review, m.total)" class="bar-warning"></mat-progress-bar>
                  <span class="bar-count">{{ m.under_review }}</span>
                </div>
                <div class="bar-item" *ngIf="m.draft">
                  <span class="bar-label muted">Draft</span>
                  <mat-progress-bar mode="determinate" [value]="pct(m.draft, m.total)" class="bar-muted"></mat-progress-bar>
                  <span class="bar-count">{{ m.draft }}</span>
                </div>
              </div>
              <div class="meeting-total">{{ m.total }} docs</div>
            </div>
            <div *ngIf="meetingSummaries.length === 0" class="empty-state">
              <mat-icon>event_busy</mat-icon><span>No meetings yet</span>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Document Aging -->
        <mat-card class="dash-card">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon red">hourglass_top</mat-icon>
            <mat-card-title>Document Aging</mat-card-title>
            <mat-card-subtitle>Active documents by age</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div *ngFor="let b of agingBuckets" class="aging-row">
              <span class="aging-label">{{ b.bucket }}</span>
              <mat-progress-bar mode="determinate"
                [value]="pct(b.count, maxAging)"
                [ngClass]="agingColor(b.bucket)">
              </mat-progress-bar>
              <span class="aging-count">{{ b.count }}</span>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Overdue Documents -->
        <mat-card class="dash-card">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon red">warning</mat-icon>
            <mat-card-title>Overdue Documents</mat-card-title>
            <mat-card-subtitle>Stalled &gt; 7 days</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div *ngIf="overdueDocuments.length === 0" class="empty-state">
              <mat-icon>check_circle_outline</mat-icon><span>No overdue documents</span>
            </div>
            <div *ngFor="let d of overdueDocuments" class="overdue-item" [routerLink]="['/documents', d.document_id]">
              <div class="overdue-main">
                <span class="overdue-title">{{ d.document_title }}</span>
                <span class="overdue-days">{{ d.days_overdue }}d overdue</span>
              </div>
              <div class="pending-meta">{{ d.meeting_title }}</div>
            </div>
          </mat-card-content>
        </mat-card>

        <!-- Recent Activity -->
        <mat-card class="dash-card span-2">
          <mat-card-header>
            <mat-icon mat-card-avatar class="card-icon blue">timeline</mat-icon>
            <mat-card-title>Recent Activity</mat-card-title>
            <mat-card-subtitle>Latest system events</mat-card-subtitle>
          </mat-card-header>
          <mat-card-content>
            <div class="timeline">
              <div *ngFor="let log of timeline" class="timeline-item">
                <div class="timeline-dot" [ngClass]="actionColor(log.action)"></div>
                <div class="timeline-content">
                  <span class="timeline-action">{{ log.action | replace:'_':' ' }}</span>
                  <span class="timeline-time">{{ log.timestamp | date:'MMM d, HH:mm' }}</span>
                </div>
              </div>
              <div *ngIf="timeline.length === 0" class="empty-state">
                <mat-icon>history</mat-icon><span>No recent activity</span>
              </div>
            </div>
          </mat-card-content>
          <mat-card-actions>
            <a routerLink="/audit" mat-button color="primary">Full Audit Log →</a>
          </mat-card-actions>
        </mat-card>

      </div>
    </div>
  `,
  styles: [`
    .page-header { display:flex; align-items:flex-start; justify-content:space-between; margin-bottom:28px; flex-wrap:wrap; gap:16px; }
    .header-actions { display:flex; gap:12px; flex-shrink:0; }

    .stat-grid { display:grid; grid-template-columns:repeat(5,1fr); gap:16px; margin-bottom:24px; }
    @media(max-width:1200px) { .stat-grid { grid-template-columns:repeat(3,1fr); } }
    @media(max-width:768px)  { .stat-grid { grid-template-columns:repeat(2,1fr); } }

    .stat-card { .stat-icon { width:40px; height:40px; border-radius:10px; display:flex; align-items:center; justify-content:center; margin-bottom:12px; mat-icon { font-size:22px; } }
      .stat-icon.pending { background:rgba(245,158,11,.15); color:var(--warning); }
      .stat-icon.review  { background:var(--blue-dim); color:var(--blue); }
      .stat-icon.overdue { background:rgba(239,68,68,.15); color:var(--danger); }
      .stat-icon.metrics { background:rgba(34,197,94,.12); color:var(--success); }
      .stat-icon.rejection { background:rgba(239,68,68,.12); color:var(--danger); }
    }

    .dashboard-grid { display:grid; grid-template-columns:repeat(2,1fr); gap:20px; }
    @media(max-width:1024px) { .dashboard-grid { grid-template-columns:1fr; } }
    .span-2 { grid-column:1/-1; }

    .dash-card {
      .card-icon { width:36px!important; height:36px!important; border-radius:10px; display:flex!important; align-items:center; justify-content:center; font-size:20px!important; color:#fff;
        &.gold   { background:linear-gradient(135deg,var(--gold),#9c7c2f); color:#000!important; }
        &.orange { background:rgba(245,158,11,.2); color:var(--warning)!important; }
        &.blue   { background:rgba(59,130,246,.2); color:var(--blue)!important; }
        &.red    { background:rgba(239,68,68,.15); color:var(--danger)!important; }
      }
    }

    .pending-item { padding:10px 0; border-bottom:1px solid var(--border); cursor:pointer; transition:padding var(--transition);
      &:hover { padding-left:6px; }
      &:last-child { border-bottom:none; }
    }
    .pending-title { font-weight:500; color:var(--text-primary); font-size:.875rem; margin-bottom:4px; }
    .pending-meta { display:flex; align-items:center; gap:4px; color:var(--text-muted); font-size:.75rem; }
    .meta-icon { font-size:14px!important; }
    .pending-date { margin-left:auto; }

    .status-chip { font-size:.7rem; padding:2px 8px; border-radius:100px; font-weight:600; display:inline-block; margin-top:4px; }

    .meeting-row { display:flex; align-items:center; gap:16px; padding:12px 0; border-bottom:1px solid var(--border); &:last-child{border:none;} }
    .meeting-info { min-width:200px; }
    .meeting-type-badge { font-size:.65rem; padding:2px 6px; background:var(--gold-dim); color:var(--gold); border-radius:4px; font-weight:700; letter-spacing:.06em; margin-right:8px; }
    .meeting-name { font-size:.875rem; color:var(--text-primary); }
    .meeting-bars { flex:1; display:flex; flex-direction:column; gap:6px; }
    .bar-item { display:flex; align-items:center; gap:8px; }
    .bar-label { font-size:.7rem; width:50px; text-align:right; &.success{color:var(--success);} &.warning{color:var(--warning);} &.muted{color:var(--text-muted);} }
    .bar-count { font-size:.75rem; color:var(--text-muted); width:20px; text-align:right; }
    .bar-success ::ng-deep .mdc-linear-progress__bar-inner { border-color:var(--success)!important; }
    .bar-warning ::ng-deep .mdc-linear-progress__bar-inner { border-color:var(--warning)!important; }
    .bar-muted   ::ng-deep .mdc-linear-progress__bar-inner { border-color:var(--text-muted)!important; }
    .meeting-total { font-size:.75rem; color:var(--text-muted); white-space:nowrap; min-width:50px; text-align:right; }

    .aging-row { display:flex; align-items:center; gap:12px; margin-bottom:12px; }
    .aging-label { font-size:.8rem; color:var(--text-secondary); width:80px; flex-shrink:0; }
    .aging-count { font-size:.8rem; font-family:'IBM Plex Mono',monospace; color:var(--text-muted); width:24px; text-align:right; }

    .overdue-item { padding:10px 0; border-bottom:1px solid var(--border); cursor:pointer; &:last-child{border:none;} }
    .overdue-main { display:flex; justify-content:space-between; align-items:center; }
    .overdue-title { font-size:.875rem; font-weight:500; color:var(--text-primary); }
    .overdue-days { font-size:.75rem; color:var(--danger); font-weight:600; }

    .timeline { display:flex; flex-direction:column; gap:12px; }
    .timeline-item { display:flex; align-items:flex-start; gap:12px; }
    .timeline-dot { width:8px; height:8px; border-radius:50%; flex-shrink:0; margin-top:4px;
      &.created  { background:var(--success); }
      &.approved { background:var(--gold); }
      &.rejected { background:var(--danger); }
      &.default  { background:var(--blue); }
    }
    .timeline-content { flex:1; display:flex; justify-content:space-between; }
    .timeline-action { font-size:.8rem; color:var(--text-primary); text-transform:capitalize; }
    .timeline-time { font-size:.75rem; color:var(--text-muted); }

    .empty-state { display:flex; flex-direction:column; align-items:center; gap:8px; padding:24px; color:var(--text-muted); font-size:.875rem;
      mat-icon { font-size:32px; width:32px; height:32px; color:var(--text-muted); opacity:.5; }
    }
  `]
})
export class DashboardComponent implements OnInit {
  private svc = inject(DashboardService);

  today = new Date();
  pendingApprovals: PendingItem[]          = [];
  pendingReviews: PendingItem[]            = [];
  overdueDocuments: OverdueDocument[]      = [];
  meetingSummaries: MeetingStatusSummary[] = [];
  agingBuckets: AgingBucket[]             = [];
  metrics: TurnaroundMetrics | null        = null;
  rejection: RejectionAnalytics | null     = null;
  timeline: AuditLog[]                     = [];

  get maxAging(): number {
    return Math.max(...this.agingBuckets.map(b => b.count), 1);
  }

  ngOnInit(): void {
    forkJoin({
      approvals:  this.svc.pendingApprovals(),
      reviews:    this.svc.pendingReviews(),
      overdue:    this.svc.overdueDocuments(),
      meetings:   this.svc.meetingStatusSummary(),
      aging:      this.svc.agingAnalysis(),
      metrics:    this.svc.turnaroundMetrics(),
      rejection:  this.svc.rejectionRatio(),
      timeline:   this.svc.activityTimeline(0, 10),
    }).subscribe(res => {
      this.pendingApprovals  = res.approvals.data  ?? [];
      this.pendingReviews    = res.reviews.data    ?? [];
      this.overdueDocuments  = res.overdue.data    ?? [];
      this.meetingSummaries  = res.meetings.data   ?? [];
      this.agingBuckets      = res.aging.data      ?? [];
      this.metrics           = res.metrics;
      this.rejection         = res.rejection;
      this.timeline          = res.timeline.data   ?? [];
    });
  }

  pct(val: number, total: number): number {
    return total === 0 ? 0 : Math.round((val / total) * 100);
  }

  agingColor(bucket: string): string {
    if (bucket.includes('15+')) return 'bar-danger';
    if (bucket.includes('8-14')) return 'bar-warning';
    return 'bar-info';
  }

  actionColor(action: string): string {
    if (action.includes('APPROVED')) return 'approved';
    if (action.includes('REJECTED')) return 'rejected';
    if (action.includes('CREATED')) return 'created';
    return 'default';
  }
}
