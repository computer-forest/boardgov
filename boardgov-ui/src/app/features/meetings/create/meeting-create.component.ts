import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatNativeDateModule } from '@angular/material/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MeetingService } from '@core/services/api.service';
import { NotificationService } from '@core/services/notification.service';

@Component({
  selector: 'bg-meeting-create',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, RouterLink,
    MatFormFieldModule, MatInputModule, MatSelectModule,
    MatDatepickerModule, MatNativeDateModule, MatButtonModule,
    MatIconModule, MatCardModule, MatProgressSpinnerModule
  ],
  template: `
    <div class="page-container">
      <div class="page-header fade-in">
        <div>
          <button mat-button routerLink="/meetings" class="back-btn">
            <mat-icon>arrow_back</mat-icon> Back to Meetings
          </button>
          <h2 class="section-title">Schedule New Meeting</h2>
          <p class="section-subtitle">Create a board or committee meeting</p>
        </div>
      </div>

      <mat-card class="form-card fade-in fade-in-delay-1">
        <mat-card-content>
          <form [formGroup]="form" (ngSubmit)="submit()">
            <div class="form-grid">
              <mat-form-field appearance="outline" class="span-2">
                <mat-label>Meeting Title</mat-label>
                <input matInput formControlName="title" placeholder="e.g. Q4 Board Meeting 2025" />
                <mat-error>Title is required</mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline">
                <mat-label>Meeting Type</mat-label>
                <mat-select formControlName="meeting_type">
                  <mat-option value="BOARD">Board</mat-option>
                  <mat-option value="COMMITTEE">Committee</mat-option>
                  <mat-option value="AGM">AGM</mat-option>
                  <mat-option value="SPECIAL">Special</mat-option>
                </mat-select>
                <mat-error>Type is required</mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline">
                <mat-label>Scheduled Date</mat-label>
                <input matInput [matDatepicker]="picker" formControlName="scheduled_at" />
                <mat-datepicker-toggle matIconSuffix [for]="picker"></mat-datepicker-toggle>
                <mat-datepicker #picker></mat-datepicker>
                <mat-error>Date is required</mat-error>
              </mat-form-field>

              <mat-form-field appearance="outline" class="span-2">
                <mat-label>Location</mat-label>
                <input matInput formControlName="location" placeholder="Boardroom A / Video Call" />
              </mat-form-field>

              <mat-form-field appearance="outline" class="span-2">
                <mat-label>Agenda</mat-label>
                <textarea matInput formControlName="agenda" rows="5"
                  placeholder="1. Financial review&#10;2. Strategy update&#10;3. AOB"></textarea>
              </mat-form-field>
            </div>

            <div class="form-actions">
              <button mat-stroked-button type="button" routerLink="/meetings">Cancel</button>
              <button mat-raised-button color="primary" type="submit" [disabled]="form.invalid || loading">
                <mat-spinner *ngIf="loading" diameter="18"></mat-spinner>
                <span *ngIf="!loading">
                  <mat-icon>event</mat-icon> Schedule Meeting
                </span>
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
    .form-grid { display:grid; grid-template-columns:1fr 1fr; gap:16px; }
    .span-2 { grid-column:1/-1; }
    .form-actions { display:flex; gap:12px; justify-content:flex-end; margin-top:8px; }
  `]
})
export class MeetingCreateComponent {
  private fb    = inject(FormBuilder);
  private svc   = inject(MeetingService);
  private nav   = inject(Router);
  private notif = inject(NotificationService);

  loading = false;

  form = this.fb.nonNullable.group({
    title:        ['', Validators.required],
    meeting_type: ['BOARD' as const, Validators.required],
    scheduled_at: ['', Validators.required],
    location:     [''],
    agenda:       ['']
  });

  submit(): void {
    if (this.form.invalid) return;
    this.loading = true;
    const v = this.form.getRawValue();
    this.svc.create({
      ...v,
      scheduled_at: new Date(v.scheduled_at).toISOString()
    } as any).subscribe({
      next: m => { this.notif.success('Meeting scheduled'); this.nav.navigate(['/meetings', m.id]); },
      error: err => { this.loading = false; this.notif.error(err.error?.error ?? 'Failed to create meeting'); }
    });
  }
}
