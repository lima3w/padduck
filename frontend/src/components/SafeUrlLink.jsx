import { isSafeHttpUrl } from '../utils/url'

// Renders a user-supplied URL as a link only when it uses http(s);
// otherwise shows it as plain text so javascript:/data: values are inert.
export default function SafeUrlLink({ value }) {
  if (!isSafeHttpUrl(value)) {
    return <span className="break-all">{value}</span>
  }
  return (
    <a href={value} target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline break-all">
      {value}
    </a>
  )
}
