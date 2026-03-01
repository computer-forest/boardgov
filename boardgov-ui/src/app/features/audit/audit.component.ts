import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatBadgeModule } from '@angular/material/badge';
import { AuditService } from '@core/services/api.service';
import { AuditLog, AuditIntegrity } from '@core/models';
import { ReplacePipe } from '@shared/pipes/replace.pipe';

@Component({
  selector: 'bg-audit',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule, MatTableModule, MatIconModule, MatButtonModule,
    MatChipsModule, MatPaginatorModule, MatProgressSpinnerModule,
    MatTooltipModule, MatBadgeModule, ReplacePipe
  ],
  template: `
    <div class="page-container">
      <!-- Header -->
      <div class="page-header fade-in">
        <div>
          <h2 class="section-title">Audit Log</h2>
          <p class="section-subtitle">Immutable, hash-chained activity record</p>
        </div>
        <button mat-stroked-button color="primary" (click)="checkIntegrity()" [disabled]="checkingIntegrity">
          <mat-icon>verified_user</mat-icon>
          {{ checkingIntegrity ? 'Validating...' : 'Validate Chain Integrity' }}
        </button>
      </div>

      <!-- Integrity Banner -->
      <div class="integrity-banner fade-in" *ngIf="integrity" [class.valid]="integrity.valid" [class.invalid]="!integrity.valid">
        <mat-icon>{{ integrity.valid ? 'verified' : 'gpp_bad' }}</mat-icon>
        <div class="integrity-text">
          <strong>{{ integrity.valid ? 'Chain Integrity Verified' : 'Chain Integrity Compromised' }}</strong>
          <span *ngIf="integrity.valid">All {{ total }} audit records form a valid tamper-proof chain.</span>
          <span *ngIf="!integrity.valid">{{ integrity.errors.length }} integrity violation(s) detected.</span>
        </div>
        <div *ngIf="!integrity.valid" class="integrity-errors">
          <div *ngFor="let e of integrity.errors" class="error-line">{{ e }}</div>
        </div>
      </div>

      <!-- Stats row -->
      <div class="stat-grid fade-in fade-in-delay-1" style="grid-template-columns:repeat(4,1fr)">
        <div class="stat-card">
          <div class="stat-value">{{ total }}</div>
          <div class="stat-label">Total Events</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ countByAction('DOCUMENT_APPROVED') }}</div>
          <div class="stat-label">Approvals</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ countByAction('DOCUMENT_REJECTED') }}</div>
          <div class="stat-label">Rejections</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ countByAction('FILE_DOWNLOADED') }}</div>
          <div class="stat-label">Downloads</div>
        </div>
      </div>

      <!-- Audit Table -->
      <mat-card class="fade-in fade-in-delay-2">
        <mat-card-content>
          <div *ngIf="loading" class="loading-center"><mat-spinner diameter="40"></mat-spinner></div>

          <table mat-table [dataSource]="logs" *ngIf="!loading" class="full-table">

            <ng-container matColumnDef="timestamp">
              <th mat-header-cell *matHeaderCellDef>Timestamp</th>
              <td mat-cell *matCellDef="let log" class="mono text-muted">
                {{ log.timestamp | date:'MMM d, y HH:mm:ss' }}
              </td>
            </ng-container>

            <ng-container matColumnDef="action">
              <th mat-header-cell *matHeaderCellDef>Action</th>
              <td mat-cell *matCellDef="let log">
                <span class="action-badge" [ngClass]="actionClass(log.action)">
                  <mat-icon>{{ actionIcon(log.action) }}</mat-icon>
                  {{ log.action | replace:'_':' ' }}
                </span>
              </td>
            </ng-container>

            <ng-container matColumnDef="transition">
              <th mat-header-cell *matHeaderCellDef>State Change</th>
              <td mat-cell *matCellDef="let log">
                <span *ngIf="log.previous_state || log.new_state" class="transition">
                  <span *ngIf="log.previous_state" class="status-chip status-{{ log.previous_state.toLowerCase() }}">
                    {{ log.previous_state }}
                  </span>
                  <mat-icon *ngIf="log.previous_state && log.new_state" class="arrow-icon">arrow_forward</mat-icon>
                  <span *ngIf="log.new_state" class="status-chip status-{{ log.new_state.toLowerCase() }}">
                    {{ log.new_state }}
                  </span>
                </span>
                <span *ngIf="!log.previous_state && !log.new_state" class="text-muted">—</span>
              </td>
            </ng-container>

            <ng-container matColumnDef="ip">
              <th mat-header-cell *matHeaderCellDef>IP Address</th>
              <td mat-cell *matCellDef="let log" class="mono text-muted">{{ log.ip_address || '—' }}</td>
            </ng-container>

            <ng-container matColumnDef="hash">
              <th mat-header-cell *matHeaderCellDef>Hash</th>
              <td mat-cell *matCellDef="let log">
                <span class="hash mono" [matTooltip]="log.hash">{{ log.hash | slice:0:12 }}…</span>
              </td>
            </ng-container>

            <tr mat-header-row *matHeaderRowDef="cols; sticky: true"></tr>
            <tr mat-row *matRowDef="let row; columns: cols;" class="audit-row"></tr>
          </table>

          <div *ngIf="!loading && logs.length === 0" class="empty-full">
            <mat-icon>history</mat-icon>
            <p>No audit records found</p>
          </div>

          <mat-paginator
            *ngIf="total > 0"
            [length]="total"
            [pageSize]="limit"
            [pageSizeOptions]="[20, 50, 100]"
            (page)="onPage($event)">
          </mat-paginator>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .page-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:24px; flex-wrap:wrap; gap:16px; }

    .integrity-banner { display:flex; align-items:flex-start; gap:16px; padding:16px 20px; border-radius:var(--radius-md); margin-bottom:20px; border:1px solid;
      mat-icon{font-size:28px!important;width:28px;height:28px;flex-shrink:0;margin-top:2px;}
      &.valid { background:rgba(34,197,94,.08); border-color:rgba(34,197,94,.25); mat-icon{color:var(--success);} }
      &.invalid { background:rgba(239,68,68,.08); border-color:rgba(239,68,68,.25); mat-icon{color:var(--danger);} }
    }
    .integrity-text { flex:1; strong{display:block;margin-bottom:4px;} span{font-size:.875rem;color:var(--text-secondary);} }
    .integrity-errors { margin-top:8px; }
    .error-line { font-size:.75rem; color:var(--danger); font-family:'IBM Plex Mono',monospace; }

    .full-table { width:100%; }
    .loading-center { display:flex; justify-content:center; padding:60px; }

    .action-badge { display:inline-flex; align-items:center; gap:5px; font-size:.7rem; padding:3px 8px; border-radius:4px; font-weight:700; letter-spacing:.04em;
      mat-icon{font-size:13px!important;width:13px;height:13px;}
      &.approved { background:rgba(34,197,94,.12); color:var(--success); }
      &.rejected  { background:rgba(239,68,68,.12); color:var(--danger); }
      &.created   { background:rgba(59,130,246,.12); color:var(--blue); }
      &.uploaded  { background:rgba(201,168,76,.12); color:var(--gold); }
      &.downloaded{ background:rgba(148,163,184,.1); color:var(--text-secondary); }
      &.default   { background:var(--bg-elevated); color:var(--text-muted); }
    }

    .transition { display:flex; align-items:center; gap:6px; flex-wrap:wrap; }
    .arrow-icon { font-size:14px!important; color:var(--text-muted); }
    .status-chip { font-size:.65rem; padding:2px 6px; border-radius:4px; font-weight:600; }

    .hash { color:var(--text-muted); font-size:.75rem; cursor:default; }
    .audit-row:hover .hash { color:var(--gold); }

    .empty-full { display:flex; flex-direction:column; align-items:center; gap:12px; padding:60px; color:var(--text-muted); mat-icon{font-size:48px;width:48px;height:48px;opacity:.3;} }
  `]
})
export class AuditComponent implements OnInit {
  private svc = inject(AuditService);

  logs: AuditLog[]          = [];
  integrity: AuditIntegrity | null = null;
  total  = 0;
  offset = 0;
  limit  = 20;
  loading = true;
  checkingIntegrity = false;

  cols = ['timestamp', 'action', 'transition', 'ip', 'hash'];

  ngOnInit(): void { this.load(); }

  load(): void {
    this.loading = true;
    this.svc.list(this.offset, this.limit).subscribe({
      next: res => { this.logs = res.data ?? []; this.total = res.total; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  onPage(e: PageEvent): void {
    this.offset = e.pageIndex * e.pageSize;
    this.limit  = e.pageSize;
    this.load();
  }

  checkIntegrity(): void {
    this.checkingIntegrity = true;
    this.svc.integrity().subscribe({
      next: res => { this.integrity = res; this.checkingIntegrity = false; },
      error: () => { this.checkingIntegrity = false; }
    });
  }

  countByAction(action: string): number {
    return this.logs.filter(l => l.action === action).length;
  }

  actionClass(action: string): string {
    if (action.includes('APPROVED')) return 'approved';
    if (action.includes('REJECTED')) return 'rejected';
    if (action.includes('CREATED') || action.includes('SUBMITTED')) return 'created';
    if (action.includes('UPLOAD')) return 'uploaded';
    if (action.includes('DOWNLOAD')) return 'downloaded';
    return 'default';
  }

  actionIcon(action: string): string {
    if (action.includes('APPROVED'))   return 'check_circle';
    if (action.includes('REJECTED'))   return 'cancel';
    if (action.includes('CREATED'))    return 'add_circle';
    if (action.includes('SUBMITTED'))  return 'send';
    if (action.includes('UPLOAD'))     return 'cloud_upload';
    if (action.includes('DOWNLOAD'))   return 'cloud_download';
    if (action.includes('REVIEW'))     return 'rate_review';
    if (action.includes('LOCKED'))     return 'lock';
    if (action.includes('ARCHIVED'))   return 'archive';
    return 'info';
  }
}
