import { Routes } from '@angular/router';

export const MEETING_ROUTES: Routes = [
  { path: '', loadComponent: () => import('./list/meeting-list.component').then(m => m.MeetingListComponent) },
  { path: 'new', loadComponent: () => import('./create/meeting-create.component').then(m => m.MeetingCreateComponent) },
  { path: ':id', loadComponent: () => import('./detail/meeting-detail.component').then(m => m.MeetingDetailComponent) }
];
