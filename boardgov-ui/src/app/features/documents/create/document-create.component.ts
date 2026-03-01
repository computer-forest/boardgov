import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink, ActivatedRoute } from '@angular/router';
import { switchMap, map, of } from 'rxjs';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatRadioModule } from '@angular/material/radio';
import { MatDividerModule } from '@angular/material/divider';
import { MatTooltipModule } from '@angular/material/tooltip';
import { DocumentService, MeetingService } from '@core/services/api.service';
import { NotificationService } from '@core/services/notification.service';
import { Meeting } from '@core/models';

const ALLOWED_TYPES = [
  'application/pdf',
  'application/zip',
  'application/x-zip-compressed',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
];
const ALLOWED_EXT = ['.pdf', '.zip', '.doc', '.docx'];
const MAX_BYTES = 30 * 1024 * 1024; // 30 MB

@Component({
  selector: 'bg-document-create',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, RouterLink,
    MatFormFieldModule, MatInputModule, MatSelectModule,
    MatButtonModule, MatIconModule, MatCardModule,
    MatProgressSpinnerModule, MatRadioModule,
    MatDividerModule, MatTooltipModule
  ],
  template: `
    <div class="page-container">
      <div class="page-header fade-in">
        <div>
          <button mat-button routerLink="/documents" class="back-btn"><mat-icon>arrow_back</mat-icon> Documents</button>
          <h2 class="section-title">New Document</h2>
          <p class="section-subtitle">Create a governance document and configure its workflow</p>
        </div>
      </div>

      <mat-card class="form-card fade-in fade-in-delay-1">
        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="submit()">
            <div class="form-grid">

              <mat-form-field appearance="outline" class="span-2">
                <mat-label>Document Title</mat-label>
                <input matInput formControlName="title" placeholder="e.g. Annual Report 2025" />
                <mat-error>Title is required</mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline" class="span-2">
                <mat-label>Description</mat-label>
                <textarea matInput formControlName="description" rows="3"
                          placeholder="Brief description of this document..."></textarea>
              </mat-form-field>

              <mat-form-field appearance="outline">
                <mat-label>Meeting</mat-label>
                <mat-select formControlName="meeting_id">
                  <mat-option *ngFor="let m of meetings" [value]="m.id">
                    {{ m.title }} ({{ m.meeting_type }})
                  </mat-option>
                </mat-select>
                <mat-error>Meeting is required</mat-error>
              </mat-form-field>

              <div class="workflow-selector">
                <label class="field-label">Workflow Mode</label>
                <mat-radio-group formControlName="workflow_mode" class="radio-group">
                  <mat-radio-button value="SEQUENTIAL">
                    <div class="radio-content">
                      <mat-icon>linear_scale</mat-icon>
                      <div>
                        <div class="radio-title">Sequential</div>
                        <div class="radio-desc">One at a time in order</div>
                      </div>
                    </div>
                  </mat-radio-button>
                  <mat-radio-button value="PARALLEL">
                    <div class="radio-content">
                      <mat-icon>call_split</mat-icon>
                      <div>
                        <div class="radio-title">Parallel</div>
                        <div class="radio-desc">All reviewers simultaneously</div>
                      </div>
                    </div>
                  </mat-radio-button>
                </mat-radio-group>
              </div>

            </div>

            <!-- ── File Upload ─────────────────────────────────── -->
            <mat-divider class="section-divider"></mat-divider>

            <div class="upload-section">
              <label class="section-label">
                <mat-icon>attach_file</mat-icon>
                Attach Document File
                <span class="optional-tag">— optional, you can also upload after creation</span>
              </label>

              <!-- Drop Zone -->
              <div class="upload-zone" *ngIf="!selectedFile"
                   [class.drag-over]="isDragging"
                   (dragover)="onDragOver($event)"
                   (dragleave)="isDragging = false"
                   (drop)="onDrop($event)"
                   (click)="fileInput.click()">
                <mat-icon class="zone-icon">cloud_upload</mat-icon>
                <div>
                  <p class="upload-title">Drop a file here or <span class="browse-link">click to browse</span></p>
                  <p class="upload-hint">PDF · ZIP · Word (.doc, .docx) &nbsp;·&nbsp; Max 30 MB</p>
                </div>
                <div class="type-chips">
                  <span class="chip pdf">PDF</span>
                  <span class="chip zip">ZIP</span>
                  <span class="chip doc">DOC</span>
                  <span class="chip doc">DOCX</span>
                </div>
              </div>

              <!-- Selected File Card -->
              <div class="file-card" *ngIf="selectedFile">
                <mat-icon class="file-icon" [class]="fileIconClass">{{ fileIcon }}</mat-icon>
                <div class="file-info">
                  <div class="file-name">{{ selectedFile.name }}</div>
                  <div class="file-meta">{{ formatSize(selectedFile.size) }} &nbsp;·&nbsp; {{ fileTypeLabel }}</div>
                </div>
                <button mat-icon-button type="button" (click)="clearFile()" matTooltip="Remove file">
                  <mat-icon>close</mat-icon>
                </button>
              </div>

              <input #fileInput type="file" accept=".pdf,.zip,.doc,.docx"
                     style="display:none" (change)="onFileSelect($event)" />
            </div>

            <!-- ── Actions ─────────────────────────────────────── -->
            <div class="form-actions">
              <button mat-stroked-button type="button" routerLink="/documents">Cancel</button>
              <button mat-raised-button color="primary" type="submit"
                      [disabled]="form.invalid || loading">
                <mat-spinner *ngIf="loading" diameter="18"></mat-spinner>
                <ng-container *ngIf="!loading">
                  <mat-icon>{{ selectedFile ? 'upload_file' : 'description' }}</mat-icon>
                  {{ selectedFile ? 'Create &amp; Upload' : 'Create Document' }}
                </ng-container>
                <span *ngIf="loading">{{ loadingMsg }}</span>
              </button>
            </div>

          </form>
        </mat-card-content>
      </mat-card>
    </div>
  `,
  styles: [`
    .page-header { margin-bottom:24px; }
    .back-btn { color:var(--text-muted); margin-bottom:8px; padding-left:0; }
    .form-card { max-width:760px; }

    .form-grid { display:grid; grid-template-columns:1fr 1fr; gap:20px; }
    .span-2 { grid-column:1/-1; }

    .field-label { font-size:.75rem; color:var(--text-secondary); display:block; margin-bottom:10px; text-transform:uppercase; letter-spacing:.06em; }
    .workflow-selector { padding:0 4px; }
    .radio-group { display:flex; gap:12px; }
    ::ng-deep .mat-mdc-radio-button { .mdc-radio { margin-right:8px; } }
    .radio-content { display:flex; align-items:center; gap:10px; mat-icon{color:var(--text-muted);} }
    .radio-title { font-size:.875rem; font-weight:500; color:var(--text-primary); }
    .radio-desc { font-size:.75rem; color:var(--text-muted); }

    .section-divider { margin:24px 0; }

    .section-label {
      display:flex; align-items:center; gap:8px;
      font-size:.75rem; font-weight:600; text-transform:uppercase; letter-spacing:.06em;
      color:var(--text-secondary); margin-bottom:14px;
      mat-icon { font-size:16px!important; width:16px; height:16px; }
    }
    .optional-tag { text-transform:none; letter-spacing:0; font-weight:400; font-size:.7rem; color:var(--text-muted); }

    /* Drop zone */
    .upload-zone {
      border:2px dashed var(--border); border-radius:var(--radius-lg);
      padding:36px 20px; text-align:center; cursor:pointer;
      transition:all var(--transition);
      display:flex; flex-direction:column; align-items:center; gap:14px;
      background:var(--bg-elevated);
    }
    .upload-zone:hover, .upload-zone.drag-over {
      border-color:var(--gold-border); background:var(--gold-dim);
    }
    .zone-icon { font-size:44px!important; width:44px; height:44px; color:var(--text-muted); opacity:.45; }
    .upload-title { font-size:.875rem; font-weight:500; color:var(--text-primary); margin:0 0 4px; }
    .browse-link { color:var(--gold); text-decoration:underline; cursor:pointer; }
    .upload-hint { font-size:.78rem; color:var(--text-muted); margin:0; }

    .type-chips { display:flex; gap:8px; flex-wrap:wrap; justify-content:center; }
    .chip { font-size:.65rem; font-weight:700; padding:2px 8px; border-radius:4px; font-family:'IBM Plex Mono',monospace; }
    .chip.pdf  { background:rgba(239,68,68,.12);  color:#ef4444; }
    .chip.zip  { background:rgba(245,158,11,.12); color:#f59e0b; }
    .chip.doc  { background:rgba(59,130,246,.12); color:#3b82f6; }

    /* Selected file card */
    .file-card {
      display:flex; align-items:center; gap:14px;
      border:1px solid var(--border); border-radius:var(--radius-lg);
      padding:14px 16px; background:var(--bg-card);
    }
    .file-icon { font-size:34px!important; width:34px; height:34px; flex-shrink:0; }
    .file-icon.pdf { color:#ef4444; }
    .file-icon.zip { color:#f59e0b; }
    .file-icon.doc { color:#3b82f6; }
    .file-info { flex:1; min-width:0; }
    .file-name { font-size:.875rem; font-weight:500; color:var(--text-primary); white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
    .file-meta { font-size:.75rem; color:var(--text-muted); margin-top:2px; }

    .form-actions { display:flex; gap:12px; justify-content:flex-end; margin-top:24px; }
    mat-spinner { display:inline-block; margin-right:8px; vertical-align:middle; }
  `]
})
export class DocumentCreateComponent implements OnInit {
  private fb      = inject(FormBuilder);
  private docSvc  = inject(DocumentService);
  private meetSvc = inject(MeetingService);
  private nav     = inject(Router);
  private route   = inject(ActivatedRoute);
  private notif   = inject(NotificationService);

  meetings: Meeting[] = [];
  loading    = false;
  loadingMsg = '';
  selectedFile: File | null = null;
  isDragging = false;

  form = this.fb.nonNullable.group({
    title:         ['', Validators.required],
    description:   [''],
    meeting_id:    ['', Validators.required],
    workflow_mode: ['SEQUENTIAL' as const, Validators.required]
  });

  ngOnInit(): void {
    this.meetSvc.list(1, 100).subscribe(res => {
      this.meetings = res.data ?? [];
      const mid = this.route.snapshot.queryParamMap.get('meetingId');
      if (mid) this.form.patchValue({ meeting_id: mid });
    });
  }

  // ── File handling ──────────────────────────────────────────────────────────

  onFileSelect(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file  = input.files?.[0];
    if (file) this.validateAndSet(file);
    input.value = '';
  }

  onDragOver(event: DragEvent): void { event.preventDefault(); this.isDragging = true; }

  onDrop(event: DragEvent): void {
    event.preventDefault();
    this.isDragging = false;
    const file = event.dataTransfer?.files?.[0];
    if (file) this.validateAndSet(file);
  }

  clearFile(): void { this.selectedFile = null; }

  private validateAndSet(file: File): void {
    if (file.size > MAX_BYTES) {
      this.notif.error('File exceeds the 30 MB limit'); return;
    }
    const ext   = '.' + (file.name.split('.').pop() ?? '').toLowerCase();
    const valid = ALLOWED_TYPES.includes(file.type) || ALLOWED_EXT.includes(ext);
    if (!valid) {
      this.notif.error('Only PDF, ZIP, or Word (.doc, .docx) files are allowed'); return;
    }
    this.selectedFile = file;
  }

  get fileIcon(): string {
    const n = this.selectedFile?.name.toLowerCase() ?? '';
    if (n.endsWith('.pdf'))                          return 'picture_as_pdf';
    if (n.endsWith('.zip'))                          return 'folder_zip';
    if (n.endsWith('.doc') || n.endsWith('.docx'))   return 'description';
    return 'insert_drive_file';
  }

  get fileIconClass(): string {
    const n = this.selectedFile?.name.toLowerCase() ?? '';
    if (n.endsWith('.pdf')) return 'pdf';
    if (n.endsWith('.zip')) return 'zip';
    return 'doc';
  }

  get fileTypeLabel(): string {
    const n = this.selectedFile?.name.toLowerCase() ?? '';
    if (n.endsWith('.pdf'))  return 'PDF Document';
    if (n.endsWith('.zip'))  return 'ZIP Archive';
    if (n.endsWith('.docx')) return 'Word Document';
    if (n.endsWith('.doc'))  return 'Word Document (Legacy)';
    return this.selectedFile?.type || 'File';
  }

  formatSize(bytes: number): string {
    if (bytes < 1024)        return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  // ── Submit ────────────────────────────────────────────────────────────────

  submit(): void {
    if (this.form.invalid) return;
    this.loading    = true;
    this.loadingMsg = 'Creating document…';

    const file = this.selectedFile;

    this.docSvc.create(this.form.getRawValue() as any).pipe(
      switchMap(doc => {
        if (!file) return of(doc);
        this.loadingMsg = 'Uploading file…';
        return this.docSvc.upload(doc.id, file).pipe(map(() => doc));
      })
    ).subscribe({
      next: d => {
        this.notif.success(file ? 'Document created and file uploaded' : 'Document created');
        this.nav.navigate(['/documents', d.id]);
      },
      error: err => {
        this.loading    = false;
        this.loadingMsg = '';
        this.notif.error(err.error?.error ?? 'Failed to create document');
      }
    });
  }
}
