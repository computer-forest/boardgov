import { Pipe, PipeTransform } from '@angular/core';

@Pipe({ name: 'replace', standalone: true, pure: true })
export class ReplacePipe implements PipeTransform {
  transform(value: string, from: string, to: string): string {
    return value ? value.split(from).join(to) : value;
  }
}
