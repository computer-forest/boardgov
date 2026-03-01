import { Injectable } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable({ providedIn: 'root' })
export class NotificationService {
  constructor(private snack: MatSnackBar) {}

  success(msg: string): void {
    this.snack.open(msg, '✕', { duration: 4000, panelClass: ['snack-success'] });
  }

  error(msg: string): void {
    this.snack.open(msg, '✕', { duration: 6000, panelClass: ['snack-error'] });
  }

  info(msg: string): void {
    this.snack.open(msg, '✕', { duration: 3000 });
  }
}
