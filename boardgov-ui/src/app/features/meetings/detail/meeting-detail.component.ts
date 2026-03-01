import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MeetingService, DocumentService } from '@core/services/api.service';
import { Meeting, Document } from '@core/models';

@Component({
  selector: 'bg-meeting-detail',
  standalone: true,
  imports: [
    CommonModule, RouterLink,
    MatCardModule, MatButtonModule, MatIconModule,
    MatChipsModule, MatDividerModule, MatProgressSpinnerModule
  ],
  template: `
    <div class="page-container" *ngIf="meeting; else loading">
      <div class="page-header fade-in">
        <div>
          <button mat-button routerLink="/meetings" class="back-btn"><mat-icon>arrow_back</mat-icon> Meetings</button>
          <div class="title-row">
            <span class="type-badge type-{{ meeting.meeting_type.toLowerCase() }}">{{ meeting.meeting_type }}</span>
            <h2 class="section-title">{{ meeting.title }}</h2>
            <mat-icon *ngIf="meeting.is_locked" class="lock-icon" matTooltip="Meeting locked">lock</mat-icon>
          </div>
          <p class="section-subtitle">
            <mat-icon class="inline-icon">event</mat-icon>{{ meeting.scheduled_at | date:'MMMM d, y — HH:mm' }}
            <span *ngIf="meeting.location"> &nbsp;·&nbsp; <mat-icon class="inline-icon">location_on</mat-icon>{{ meeting.location }}</span>
          </p>
        </div>
        <button mat-raised-button color="primary" (click)="createDocument()">
          <mat-icon>add</mat-icon> Add Document
        </button>
      </div>

      <div class="detail-grid">
        <mat-card class="fade-in fade-in-delay-1">
          <mat-card-header><mat-card-title>Agenda</mat-card-title></mat-card-header>
          <mat-card-content>
            <pre class="agenda-text">{{ meeting.agenda || 'No agenda set.' }}</pre>
          </mat-card-content>
        </mat-card>

        <mat-card class="fade-in fade-in-delay-2">
          <mat-card-header>
            <mat-card-title>Documents ({{ documents.length }})</mat-card-title>
          </mat-card-header>
          <mat-card-content>
            <div *ngIf="documents.length === 0" class="empty-state">
              <mat-icon>description</mat-icon>
              <span>No documents yet</span>
            </div>
            <div *ngFor="let doc of documents" class="doc-item" [routerLink]="['/documents', doc.id]">
              <div class="doc-info">
                <mat-icon class="doc-icon">description</mat-icon>
                <div>
                  <div class="doc-title">{{ doc.title }}</div>
                  <div class="doc-meta">{{ doc.workflow_mode }} · {{ doc.created_at | date:'MMM d' }}</div>
                </div>
              </div>
              <span class="status-chip status-{{ doc.status.toLowerCase().replace('_', '-') }}">{{ doc.status }}</span>
            </div>
          </mat-card-content>
        </mat-card>
      </div>
    </div>

    <ng-template #loading>
      <div class="loading-center"><mat-spinner></mat-spinner></div>
    </ng-template>
  `,
  styles: [`
    .page-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:24px; flex-wrap:wrap; gap:16px; }
    .back-btn { color:var(--text-muted); margin-bottom:8px; padding-left:0; }
    .title-row { display:flex; align-items:center; gap:12px; margin-bottom:6px; }
    .lock-icon { color:var(--warning); font-size:20px!important; }
    .inline-icon { font-size:14px!important; vertical-align:middle; color:var(--text-muted); }
    .type-badge { font-size:.7rem; padding:3px 8px; border-radius:4px; font-weight:700; letter-spacing:.06em; flex-shrink:0; }
    .type-board { background:rgba(201,168,76,.15); color:var(--gold); }
    .type-committee { background:rgba(59,130,246,.15); color:var(--blue); }
    .type-agm { background:rgba(34,197,94,.12); color:var(--success); }
    .type-special { background:rgba(239,68,68,.12); color:var(--danger); }
    .detail-grid { display:grid; grid-template-columns:1fr 2fr; gap:20px; }
    @media(max-width:900px) { .detail-grid { grid-template-columns:1fr; } }
    .agenda-text { font-family:'IBM Plex Mono',monospace; font-size:.8rem; color:var(--text-secondary); white-space:pre-wrap; line-height:1.8; }
    .doc-item { display:flex; align-items:center; justify-content:space-between; padding:10px 0; border-bottom:1px solid var(--border); cursor:pointer; transition:padding var(--transition); &:hover{padding-left:6px;} &:last-child{border:none;} }
    .doc-info { display:flex; align-items:center; gap:10px; }
    .doc-icon { color:var(--text-muted); }
    .doc-title { font-size:.875rem; font-weight:500; color:var(--text-primary); }
    .doc-meta { font-size:.75rem; color:var(--text-muted); }
    .status-chip { font-size:.7rem; padding:2px 8px; border-radius:100px; font-weight:600; }
    .empty-state { display:flex; flex-direction:column; align-items:center; gap:8px; padding:32px; color:var(--text-muted); mat-icon{font-size:32px;width:32px;height:32px;opacity:.3;} }
    .loading-center { display:flex; justify-content:center; padding:80px; }
  `]
})
export class MeetingDetailComponent implements OnInit {
  private route   = inject(ActivatedRoute);
  private meetSvc = inject(MeetingService);
  private docSvc  = inject(DocumentService);

  meeting: Meeting | null = null;
  documents: Document[]   = [];

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id')!;
    this.meetSvc.getById(id).subscribe(m => {
      this.meeting = m;
      this.docSvc.listByMeeting(id).subscribe(docs => { this.documents = docs ?? []; });
    });
  }

  createDocument(): void {
    // Navigate to document create with meetingId pre-filled
    window.location.href = `/documents/new?meetingId=${this.meeting?.id}`;
  }
}
