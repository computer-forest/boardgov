import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { AuthService } from '@core/auth/auth.service';
import { NotificationService } from '@core/services/notification.service';

@Component({
  selector: 'bg-login',
  standalone: true,
  imports: [
    CommonModule, ReactiveFormsModule, RouterLink,
    MatFormFieldModule, MatInputModule, MatButtonModule,
    MatIconModule, MatProgressSpinnerModule
  ],
  template: `
    <div class="auth-wrapper">
      <!-- Background decoration -->
      <div class="bg-decoration">
        <div class="hex hex-1"></div>
        <div class="hex hex-2"></div>
        <div class="hex hex-3"></div>
      </div>

      <div class="auth-card fade-in">
        <!-- Header -->
        <div class="auth-header">
          <div class="auth-logo">
            <span>⬡</span>
          </div>
          <h1 class="auth-title">BoardGov</h1>
          <p class="auth-subtitle">Governance Portal — Secure Access</p>
        </div>

        <!-- Form -->
        <form [formGroup]="form" (ngSubmit)="submit()" class="auth-form" autocomplete="off">
          <mat-form-field appearance="outline" class="full-width">
            <mat-label>Email Address</mat-label>
            <input matInput type="email" formControlName="email" placeholder="admin@company.com" autocomplete="email" />
            <mat-icon matPrefix>alternate_email</mat-icon>
            <mat-error *ngIf="form.get('email')?.hasError('required')">Email is required</mat-error>
            <mat-error *ngIf="form.get('email')?.hasError('email')">Invalid email format</mat-error>
          </mat-form-field>

          <mat-form-field appearance="outline" class="full-width">
            <mat-label>Password</mat-label>
            <input matInput [type]="showPass ? 'text' : 'password'" formControlName="password" autocomplete="current-password" />
            <mat-icon matPrefix>lock_outline</mat-icon>
            <button mat-icon-button matSuffix type="button" (click)="showPass = !showPass">
              <mat-icon>{{ showPass ? 'visibility_off' : 'visibility' }}</mat-icon>
            </button>
            <mat-error *ngIf="form.get('password')?.hasError('required')">Password is required</mat-error>
          </mat-form-field>

          <button mat-raised-button color="primary" type="submit"
                  class="submit-btn" [disabled]="form.invalid || loading">
            <mat-spinner *ngIf="loading" diameter="18" class="btn-spinner"></mat-spinner>
            <span *ngIf="!loading">Sign In</span>
          </button>

          <p class="auth-footer">
            Don't have an account?
            <a routerLink="/auth/register" class="auth-link">Create one</a>
          </p>
        </form>
      </div>
    </div>
  `,
  styles: [`
    .auth-wrapper {
      min-height: 100vh;
      background: var(--bg-base);
      display: flex; align-items: center; justify-content: center;
      position: relative; overflow: hidden;
    }

    .bg-decoration { position: absolute; inset: 0; pointer-events: none; }
    .hex {
      position: absolute;
      border: 1px solid var(--gold-border);
      opacity: 0.3;
      border-radius: 4px;
      animation: rotateHex 20s linear infinite;
    }
    .hex-1 { width: 600px; height: 600px; top: -200px; right: -200px; animation-duration: 25s; }
    .hex-2 { width: 400px; height: 400px; bottom: -150px; left: -100px; animation-duration: 35s; animation-direction: reverse; }
    .hex-3 { width: 200px; height: 200px; top: 30%; left: 10%; animation-duration: 18s; }

    @keyframes rotateHex {
      from { transform: rotate(0deg); }
      to   { transform: rotate(360deg); }
    }

    .auth-card {
      width: 100%;
      max-width: 420px;
      background: var(--bg-card);
      border: 1px solid var(--gold-border);
      border-radius: var(--radius-xl);
      padding: 48px 40px;
      box-shadow: var(--shadow-panel), var(--shadow-glow);
      position: relative; z-index: 1;
    }

    .auth-header { text-align: center; margin-bottom: 36px; }
    .auth-logo {
      width: 56px; height: 56px;
      background: linear-gradient(135deg, var(--gold), #9c7c2f);
      border-radius: 16px;
      display: flex; align-items: center; justify-content: center;
      font-size: 26px; color: #000;
      margin: 0 auto 16px;
      box-shadow: 0 8px 24px rgba(201,168,76,0.35);
    }
    .auth-title {
      font-family: 'Playfair Display', serif;
      font-size: 1.75rem; font-weight: 700;
      color: var(--text-primary); margin-bottom: 6px;
    }
    .auth-subtitle {
      font-size: 0.8rem;
      color: var(--text-muted);
      letter-spacing: 0.06em;
      text-transform: uppercase;
    }

    .auth-form { display: flex; flex-direction: column; gap: 16px; }
    .full-width { width: 100%; }

    .submit-btn {
      width: 100%; height: 48px; font-size: 0.95rem;
      font-weight: 600; letter-spacing: 0.04em;
      margin-top: 8px;
      display: flex; align-items: center; justify-content: center; gap: 8px;
    }
    .btn-spinner { --mdc-circular-progress-active-indicator-color: #000; }

    .auth-footer {
      text-align: center; font-size: 0.85rem; color: var(--text-muted);
      margin-top: 4px;
    }
    .auth-link { color: var(--gold); text-decoration: none; font-weight: 500; }
    .auth-link:hover { text-decoration: underline; }
  `]
})
export class LoginComponent {
  private fb   = inject(FormBuilder);
  private auth = inject(AuthService);
  private nav  = inject(Router);
  private notif = inject(NotificationService);

  showPass = false;
  loading  = false;

  form = this.fb.nonNullable.group({
    email:    ['', [Validators.required, Validators.email]],
    password: ['', Validators.required]
  });

  submit(): void {
    debugger;
    if (this.form.invalid) return;
    this.loading = true;
    const { email, password } = this.form.getRawValue();

    this.auth.login(email, password).subscribe({
      next: () => { this.nav.navigate(['/dashboard']); },
      error: (err) => {
        this.loading = false;
        this.notif.error(err.error?.error ?? 'Invalid credentials');
      }
    });
  }
}
