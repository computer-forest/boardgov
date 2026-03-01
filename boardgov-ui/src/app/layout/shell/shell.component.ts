import { Component, computed, inject } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { CommonModule } from '@angular/common';
import { AuthService } from '@core/auth/auth.service';

interface NavItem {
  icon: string;
  label: string;
  path: string;
  badge?: number;
}

@Component({
  selector: 'bg-shell',
  standalone: true,
  imports: [
    CommonModule, RouterOutlet, RouterLink, RouterLinkActive,
    MatSidenavModule, MatToolbarModule, MatIconModule,
    MatButtonModule, MatTooltipModule, MatMenuModule, MatDividerModule
  ],
  template: `
    <mat-sidenav-container class="shell-container">
      <!-- Sidenav -->
      <mat-sidenav mode="side" opened class="sidenav">
        <!-- Logo -->
        <div class="sidenav-logo">
          <div class="logo-mark">
            <span class="logo-icon">⬡</span>
          </div>
          <div class="logo-text">
            <span class="logo-name">BoardGov</span>
            <span class="logo-sub">Governance Portal</span>
          </div>
        </div>

        <div class="sidenav-divider"></div>

        <!-- Navigation -->
        <nav class="sidenav-nav">
          <a *ngFor="let item of navItems"
             [routerLink]="item.path"
             routerLinkActive="active"
             class="nav-item"
             [matTooltip]="item.label"
             matTooltipPosition="right">
            <mat-icon class="nav-icon">{{ item.icon }}</mat-icon>
            <span class="nav-label">{{ item.label }}</span>
            <span *ngIf="item.badge" class="nav-badge">{{ item.badge }}</span>
          </a>
        </nav>

        <div class="sidenav-spacer"></div>

        <!-- User footer -->
        <div class="sidenav-user" [matMenuTriggerFor]="userMenu">
          <div class="user-avatar">{{ userInitial() }}</div>
          <div class="user-info">
            <span class="user-name">{{ user()?.full_name }}</span>
            <span class="user-email">{{ user()?.email }}</span>
          </div>
          <mat-icon class="user-chevron">expand_less</mat-icon>
        </div>

        <mat-menu #userMenu="matMenu" yPosition="above">
          <button mat-menu-item (click)="logout()">
            <mat-icon>logout</mat-icon>
            <span>Sign Out</span>
          </button>
        </mat-menu>
      </mat-sidenav>

      <!-- Main content -->
      <mat-sidenav-content class="main-content">
        <router-outlet />
      </mat-sidenav-content>
    </mat-sidenav-container>
  `,
  styles: [`
    .shell-container { height: 100vh; }

    .sidenav {
      width: 240px;
      display: flex;
      flex-direction: column;
      padding: 0;
    }

    .sidenav-logo {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 24px 20px 20px;
    }
    .logo-mark {
      width: 36px; height: 36px;
      background: linear-gradient(135deg, var(--gold) 0%, #E8C96A 100%);
      border-radius: 10px;
      display: flex; align-items: center; justify-content: center;
      font-size: 18px; color: #000; font-weight: bold;
      box-shadow: 0 4px 12px rgba(201,168,76,0.35);
    }
    .logo-name {
      display: block;
      font-family: 'Playfair Display', serif;
      font-size: 1.1rem; font-weight: 700;
      color: var(--text-primary);
    }
    .logo-sub {
      display: block;
      font-size: 0.65rem; color: var(--text-muted);
      letter-spacing: 0.08em; text-transform: uppercase;
    }

    .sidenav-divider {
      height: 1px; background: var(--border); margin: 0 16px 16px;
    }

    .sidenav-nav {
      display: flex; flex-direction: column; gap: 2px; padding: 0 10px;
    }

    .nav-item {
      display: flex; align-items: center; gap: 12px;
      padding: 10px 12px;
      border-radius: var(--radius-md);
      color: var(--text-secondary);
      text-decoration: none;
      font-size: 0.875rem; font-weight: 500;
      transition: all var(--transition);
      position: relative;

      &:hover {
        background: var(--bg-elevated);
        color: var(--text-primary);
      }

      &.active {
        background: var(--gold-dim);
        color: var(--gold);
        .nav-icon { color: var(--gold); }
      }
    }

    .nav-icon { font-size: 20px !important; width: 20px; height: 20px; }
    .nav-badge {
      margin-left: auto;
      background: var(--gold);
      color: #000;
      font-size: 10px; font-weight: 700;
      padding: 1px 6px; border-radius: 100px;
    }

    .sidenav-spacer { flex: 1; }

    .sidenav-user {
      display: flex; align-items: center; gap: 10px;
      padding: 16px 16px 20px;
      cursor: pointer; border-top: 1px solid var(--border);
      transition: background var(--transition);

      &:hover { background: var(--bg-elevated); }
    }

    .user-avatar {
      width: 34px; height: 34px;
      background: linear-gradient(135deg, var(--gold), #9c7c2f);
      border-radius: 50%; display: flex;
      align-items: center; justify-content: center;
      font-weight: 700; font-size: 14px; color: #000;
      flex-shrink: 0;
    }
    .user-name  { display: block; font-size: 0.8rem; font-weight: 600; color: var(--text-primary); }
    .user-email { display: block; font-size: 0.7rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 130px; }
    .user-chevron { color: var(--text-muted) !important; font-size: 18px !important; margin-left: auto; }

    .main-content { background: var(--bg-base); overflow-y: auto; }
  `]
})
export class ShellComponent {
  private auth = inject(AuthService);
  readonly user        = this.auth.user;
  readonly userInitial = computed(() => this.user()?.full_name?.charAt(0).toUpperCase() ?? 'U');

  navItems: NavItem[] = [
    { icon: 'dashboard',        label: 'Dashboard',  path: '/dashboard' },
    { icon: 'event',            label: 'Meetings',   path: '/meetings'  },
    { icon: 'description',      label: 'Documents',  path: '/documents' },
    { icon: 'verified_user',    label: 'Audit Log',  path: '/audit'     },
  ];

  logout(): void { this.auth.logout(); }
}
