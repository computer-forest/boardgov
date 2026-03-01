import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatTableModule } from '@angular/material/table';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatChipsModule } from '@angular/material/chips';
import { MatPaginatorModule, PageEvent } from '@angular/material/paginator';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MeetingService } from '@core/services/api.service';
import { Meeting } from '@core/models';

@Component({
  selector: 'bg-meeting-list',
  standalone: true,
  imports: [
    CommonModule, RouterLink,
    MatTableModule, MatButtonModule, MatIconModule,
    MatCardModule, MatChipsModule, MatPaginatorModule, MatProgressSpinnerModule
  ],
  template: `
    <div class="page-container">
      <div class="page-header fade-in">
        <div>
          <h2 class="section-title">Meetings</h2>
          <p class="section-subtitle">{{ total }} meetings on record</p>
        </div>
        <button mat-raised-button color="primary" routerLink="/meetings/new">
          <mat-icon>add</mat-icon> Schedule Meeting
        </button>
      </div>

      <mat-card class="fade-in fade-in-delay-1">
        <mat-card-content>
          <div *ngIf="loading" class="loading-center">
            <mat-spinner diameter="40"></mat-spinner>
          </div>

          <table mat-table [dataSource]="meetings" *ngIf="!loading" class="full-table">
            <ng-container matColumnDef="type">
              <th mat-header-cell *matHeaderCellDef>Type</th>
              <td mat-cell *matCellDef="let m">
                <span class="type-badge type-{{ m.meeting_type.toLowerCase() }}">{{ m.meeting_type }}</span>
              </td>
            </ng-container>

            <ng-container matColumnDef="title">
              <th mat-header-cell *matHeaderCellDef>Title</th>
              <td mat-cell *matCellDef="let m">
                <a [routerLink]="['/meetings', m.id]" class="table-link">{{ m.title }}</a>
              </td>
            </ng-container>

            <ng-container matColumnDef="scheduled">
              <th mat-header-cell *matHeaderCellDef>Scheduled</th>
              <td mat-cell *matCellDef="let m" class="mono text-muted">
                {{ m.scheduled_at | date:'MMM d, y HH:mm' }}
              </td>
            </ng-container>

            <ng-container matColumnDef="location">
              <th mat-header-cell *matHeaderCellDef>Location</th>
              <td mat-cell *matCellDef="let m" class="text-muted">{{ m.location || '—' }}</td>
            </ng-container>

            <ng-container matColumnDef="locked">
              <th mat-header-cell *matHeaderCellDef>Status</th>
              <td mat-cell *matCellDef="let m">
                <span class="lock-badge" [class.locked]="m.is_locked">
                  <mat-icon>{{ m.is_locked ? 'lock' : 'lock_open' }}</mat-icon>
                  {{ m.is_locked ? 'Locked' : 'Active' }}
                </span>
              </td>
            </ng-container>

            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let m">
                <button mat-icon-button [routerLink]="['/meetings', m.id]" matTooltip="View details">
                  <mat-icon>chevron_right</mat-icon>
                </button>
              </td>
            </ng-container>

            <tr mat-header-row *matHeaderRowDef="cols"></tr>
            <tr mat-row *matRowDef="let row; columns: cols;"></tr>
          </table>

          <div *ngIf="!loading && meetings.length === 0" class="empty-full">
            <mat-icon>event_busy</mat-icon>
            <p>No meetings scheduled yet</p>
            <button mat-raised-button color="primary" routerLink="/meetings/new">Schedule First Meeting</button>
          </div>

          <mat-paginator
            *ngIf="total > 0"
            [length]="total"
            [pageSize]="limit"
            [pageSizeOptions]="[10, 20, 50]"
            (page)="onPage($event)">
          </mat-paginator>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .page-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:24px; }
    .full-table { width:100%; }
    .loading-center { display:flex; justify-content:center; padding:60px; }
    .table-link { color:var(--gold); text-decoration:none; font-weight:500; &:hover{text-decoration:underline;} }
    .type-badge { font-size:.7rem; padding:3px 8px; border-radius:4px; font-weight:700; letter-spacing:.06em; }
    .type-board { background:rgba(201,168,76,.15); color:var(--gold); }
    .type-committee { background:rgba(59,130,246,.15); color:var(--blue); }
    .type-agm { background:rgba(34,197,94,.12); color:var(--success); }
    .type-special { background:rgba(239,68,68,.12); color:var(--danger); }
    .lock-badge { display:flex; align-items:center; gap:4px; font-size:.75rem; color:var(--text-muted); mat-icon{font-size:14px!important;width:14px;height:14px;} &.locked{color:var(--warning);} }
    .empty-full { display:flex; flex-direction:column; align-items:center; gap:16px; padding:60px; color:var(--text-muted); mat-icon{font-size:48px;width:48px;height:48px;opacity:.3;} }
  `]
})
export class MeetingListComponent implements OnInit {
  private svc = inject(MeetingService);

  meetings: Meeting[] = [];
  total   = 0;
  offset  = 0;
  limit   = 20;
  loading = true;
  cols    = ['type', 'title', 'scheduled', 'location', 'locked', 'actions'];

  ngOnInit(): void { this.load(); }

  load(): void {
    this.loading = true;
    this.svc.list(this.offset, this.limit).subscribe({
      next: res => { this.meetings = res.data ?? []; this.total = res.total; this.loading = false; },
      error: () => { this.loading = false; }
    });
  }

  onPage(e: PageEvent): void {
    this.offset = e.pageIndex * e.pageSize;
    this.limit  = e.pageSize;
    this.load();
  }
}
