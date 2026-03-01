import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { tap, catchError } from 'rxjs/operators';
import { Observable, throwError } from 'rxjs';
import { environment } from '@env/environment';
import { TokenPair, User } from '@core/models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly ACCESS_TOKEN_KEY  = 'bg_access_token';
  private readonly REFRESH_TOKEN_KEY = 'bg_refresh_token';
  private readonly USER_KEY          = 'bg_user';

  private _user = signal<User | null>(this.loadUser());
  readonly user    = this._user.asReadonly();
  readonly isAuth  = computed(() => !!this._user());

  constructor(private http: HttpClient, private router: Router) {}

  login(email: string, password: string): Observable<TokenPair> {
    return this.http.post<TokenPair>(`${environment.apiUrl}/auth/login`, { email, password }).pipe(
      tap(res => this.handleTokens(res)),
      catchError(err => throwError(() => err))
    );
  }

  register(email: string, password: string, full_name: string): Observable<User> {
    return this.http.post<User>(`${environment.apiUrl}/auth/register`, { email, password, full_name });
  }

  logout(): void {
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_TOKEN_KEY);
    localStorage.removeItem(this.USER_KEY);
    this._user.set(null);
    this.router.navigate(['/auth/login']);
  }

  refreshToken(): Observable<TokenPair> {
    const refresh_token = this.getRefreshToken();
    return this.http.post<TokenPair>(`${environment.apiUrl}/auth/refresh`, { refresh_token }).pipe(
      tap(res => {
        localStorage.setItem(this.ACCESS_TOKEN_KEY, res.access_token);
        localStorage.setItem(this.REFRESH_TOKEN_KEY, res.refresh_token);
      })
    );
  }

  getAccessToken(): string | null {
    return localStorage.getItem(this.ACCESS_TOKEN_KEY);
  }

  getRefreshToken(): string | null {
    return localStorage.getItem(this.REFRESH_TOKEN_KEY);
  }

  private handleTokens(res: TokenPair): void {
    console.log(res);
    
    localStorage.setItem(this.ACCESS_TOKEN_KEY, res.access_token);
    localStorage.setItem(this.REFRESH_TOKEN_KEY, res.refresh_token);
    localStorage.setItem(this.USER_KEY, JSON.stringify(res.user));
    this._user.set(res.user);
  }

  private loadUser(): User | null {
    try {
      const raw = localStorage.getItem(this.USER_KEY);
      return raw ? JSON.parse(raw) : null;
    } catch { return null; }
  }
}
