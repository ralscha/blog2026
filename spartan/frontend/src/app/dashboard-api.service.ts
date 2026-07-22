import { HttpClient } from '@angular/common/http';
import { Service, inject, signal } from '@angular/core';

export interface DashboardMetric {
  label: string;
  value: string;
  change: string;
  trend: 'up' | 'down';
}

export interface Project {
  id: string;
  name: string;
  owner: string;
  status: string;
  completion: number;
}

export interface Activity {
  id: string;
  initials: string;
  person: string;
  action: string;
  when: string;
}

export interface DashboardData {
  metrics: DashboardMetric[];
  projects: Project[];
  activity: Activity[];
  generatedAt: string;
}

@Service()
export class DashboardApi {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = 'http://localhost:8080/api';

  readonly dashboard = signal<DashboardData | null>(null);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  load(): void {
    this.loading.set(true);
    this.error.set(null);
    this.http.get<DashboardData>(`${this.baseUrl}/dashboard`).subscribe({
      next: (data) => this.dashboard.set(data),
      error: () => this.error.set('The Go API is offline. Start it with go run . from /backend.'),
      complete: () => this.loading.set(false),
    });
  }

  createProject(name: string, completed: () => void): void {
    this.http.post<Project>(`${this.baseUrl}/projects`, { name }).subscribe({
      next: () => {
        this.load();
        completed();
      },
      error: () => this.error.set('Could not create the project. Check that the API is running.'),
    });
  }

  saveDigest(dailyDigest: boolean, completed: () => void): void {
    this.http.patch(`${this.baseUrl}/preferences`, { dailyDigest }).subscribe({
      next: completed,
      error: () => this.error.set('Could not save notification preferences.'),
    });
  }
}
