import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatInputModule } from '@angular/material/input';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSelectModule } from '@angular/material/select';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTooltipModule } from '@angular/material/tooltip';
import { DashboardService } from '@core/services/api.service';
import { PendingItem, DocumentStatus } from '@core/models';

@Component({
  selector: 'bg-document-list',
  standalone: true,
  imports: [
    CommonModule, RouterLink, ReactiveFormsModule,
    MatTableModule, MatButtonModule, MatIconModule, MatCardModule,
    MatInputModule, MatFormFieldModule, MatSelectModule,
    MatProgressSpinnerModule, MatTooltipModule
  ],
  template: `
    <div class="page-container">
      <div class="page-header fade-in">
        <div>
          <h2 class="section-title">Documents</h2>
          <p class="section-subtitle">Board & committee documentation workflow</p>
        </div>
        <button mat-raised-button color="primary" routerLink="/documents/new">
          <mat-icon>add</mat-icon> New Document
        </button>
      </div>

      <!-- Tabs -->
      <div class="doc-tabs fade-in">
        <button *ngFor="let tab of tabs" class="doc-tab"
                [class.active]="activeTab === tab.key"
                (click)="activeTab = tab.key">
          {{ tab.label }}
          <span class="tab-count" *ngIf="tab.count > 0">{{ tab.count }}</span>
        </button>
      </div>

      <!-- Pending Approvals table (My Queue) -->
      <mat-card class="fade-in fade-in-delay-1" *ngIf="activeTab === 'approvals'">
        <mat-card-content>
          <div *ngIf="loading" class="loading-center"><mat-spinner diameter="40"></mat-spinner></div>
          <table mat-table [dataSource]="pendingApprovals" *ngIf="!loading">
            <ng-container matColumnDef="title">
              <th mat-header-cell *matHeaderCellDef>Document</th>
              <td mat-cell *matCellDef="let d">
                <a [routerLink]="['/documents', d.document_id]" class="table-link">{{ d.document_title }}</a>
              </td>
            </ng-container>
            <ng-container matColumnDef="meeting">
              <th mat-header-cell *matHeaderCellDef>Meeting</th>
              <td mat-cell *matCellDef="let d" class="text-muted">{{ d.meeting_title }}</td>
            </ng-container>
            <ng-container matColumnDef="status">
              <th mat-header-cell *matHeaderCellDef>Status</th>
              <td mat-cell *matCellDef="let d">
                <span class="status-chip status-{{ d.status.toLowerCase() }}">{{ d.status }}</span>
              </td>
            </ng-container>
            <ng-container matColumnDef="assigned">
              <th mat-header-cell *matHeaderCellDef>Assigned</th>
              <td mat-cell *matCellDef="let d" class="mono text-muted">{{ d.assigned_at | date:'MMM d' }}</td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let d">
                <a mat-stroked-button color="primary" [routerLink]="['/documents', d.document_id]">
                  Review <mat-icon>chevron_right</mat-icon>
                </a>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="approvalCols"></tr>
            <tr mat-row *matRowDef="let row; columns: approvalCols;"></tr>
          </table>
          <div *ngIf="!loading && pendingApprovals.length === 0" class="empty-full">
            <mat-icon>check_circle_outline</mat-icon>
            <p>No pending approvals</p>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Pending Reviews -->
      <mat-card class="fade-in fade-in-delay-1" *ngIf="activeTab === 'reviews'">
        <mat-card-content>
          <div *ngIf="loading" class="loading-center"><mat-spinner diameter="40"></mat-spinner></div>
          <table mat-table [dataSource]="pendingReviews" *ngIf="!loading">
            <ng-container matColumnDef="title">
              <th mat-header-cell *matHeaderCellDef>Document</th>
              <td mat-cell *matCellDef="let d">
                <a [routerLink]="['/documents', d.document_id]" class="table-link">{{ d.document_title }}</a>
              </td>
            </ng-container>
            <ng-container matColumnDef="meeting">
              <th mat-header-cell *matHeaderCellDef>Meeting</th>
              <td mat-cell *matCellDef="let d" class="text-muted">{{ d.meeting_title }}</td>
            </ng-container>
            <ng-container matColumnDef="assigned">
              <th mat-header-cell *matHeaderCellDef>Assigned</th>
              <td mat-cell *matCellDef="let d" class="mono text-muted">{{ d.assigned_at | date:'MMM d' }}</td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let d">
                <a mat-stroked-button color="primary" [routerLink]="['/documents', d.document_id]">
                  Review <mat-icon>chevron_right</mat-icon>
                </a>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="reviewCols"></tr>
            <tr mat-row *matRowDef="let row; columns: reviewCols;"></tr>
          </table>
          <div *ngIf="!loading && pendingReviews.length === 0" class="empty-full">
            <mat-icon>check_circle_outline</mat-icon><p>No pending reviews</p>
          </div>
        </mat-card-content>
      </mat-card>

      <!-- Overdue -->
      <mat-card class="fade-in fade-in-delay-1" *ngIf="activeTab === 'overdue'">
        <mat-card-content>
          <div *ngIf="loading" class="loading-center"><mat-spinner diameter="40"></mat-spinner></div>
          <table mat-table [dataSource]="overdueItems" *ngIf="!loading">
            <ng-container matColumnDef="title">
              <th mat-header-cell *matHeaderCellDef>Document</th>
              <td mat-cell *matCellDef="let d">
                <a [routerLink]="['/documents', d.document_id]" class="table-link">{{ d.document_title }}</a>
              </td>
            </ng-container>
            <ng-container matColumnDef="meeting">
              <th mat-header-cell *matHeaderCellDef>Meeting</th>
              <td mat-cell *matCellDef="let d" class="text-muted">{{ d.meeting_title }}</td>
            </ng-container>
            <ng-container matColumnDef="status">
              <th mat-header-cell *matHeaderCellDef>Status</th>
              <td mat-cell *matCellDef="let d">
                <span class="status-chip status-{{ d.status.toLowerCase() }}">{{ d.status }}</span>
              </td>
            </ng-container>
            <ng-container matColumnDef="overdue">
              <th mat-header-cell *matHeaderCellDef>Days Overdue</th>
              <td mat-cell *matCellDef="let d" class="text-danger mono">{{ d.days_overdue }}d</td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="overdueCols"></tr>
            <tr mat-row *matRowDef="let row; columns: overdueCols;"></tr>
          </table>
          <div *ngIf="!loading && overdueItems.length === 0" class="empty-full">
            <mat-icon>check_circle_outline</mat-icon><p>No overdue documents</p>
          </div>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .page-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:24px; flex-wrap:wrap; gap:16px; }
    .doc-tabs { display:flex; gap:4px; background:var(--bg-surface); border:1px solid var(--border); border-radius:var(--radius-md); padding:4px; margin-bottom:20px; width:fit-content; }
    .doc-tab { background:transparent; border:none; color:var(--text-muted); padding:8px 18px; border-radius:var(--radius-sm); cursor:pointer; font-size:.875rem; font-weight:500; transition:all var(--transition); display:flex; align-items:center; gap:6px;
      &:hover{color:var(--text-primary);}
      &.active{background:var(--bg-elevated); color:var(--gold);}
    }
    .tab-count { background:var(--danger); color:#fff; font-size:.65rem; padding:1px 5px; border-radius:100px; font-weight:700; }
    .table-link { color:var(--gold); text-decoration:none; font-weight:500; &:hover{text-decoration:underline;} }
    .status-chip { font-size:.7rem; padding:2px 8px; border-radius:100px; font-weight:600; }
    .loading-center { display:flex; justify-content:center; padding:60px; }
    .empty-full { display:flex; flex-direction:column; align-items:center; gap:12px; padding:60px; color:var(--text-muted); mat-icon{font-size:40px;width:40px;height:40px;opacity:.3;} }
  `]
})
export class DocumentListComponent implements OnInit {
  private svc = inject(DashboardService);

  activeTab = 'approvals';
  loading   = true;

  pendingApprovals: any[] = [];
  pendingReviews: any[]   = [];
  overdueItems: any[]     = [];

  approvalCols = ['title', 'meeting', 'status', 'assigned', 'actions'];
  reviewCols   = ['title', 'meeting', 'assigned', 'actions'];
  overdueCols  = ['title', 'meeting', 'status', 'overdue'];

  get tabs() {
    return [
      { key: 'approvals', label: 'Pending Approvals', count: this.pendingApprovals.length },
      { key: 'reviews',   label: 'Pending Reviews',   count: this.pendingReviews.length },
      { key: 'overdue',   label: 'Overdue',            count: this.overdueItems.length },
    ];
  }

  ngOnInit(): void {
    this.loading = true;
    Promise.all([
      this.svc.pendingApprovals().toPromise(),
      this.svc.pendingReviews().toPromise(),
      this.svc.overdueDocuments().toPromise()
    ]).then(([a, r, o]) => {
      this.pendingApprovals = a?.data ?? [];
      this.pendingReviews   = r?.data ?? [];
      this.overdueItems     = o?.data ?? [];
      this.loading = false;
    }).catch(() => { this.loading = false; });
  }
}
