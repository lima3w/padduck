const PORT_NAMES = {
  21: 'FTP', 22: 'SSH', 23: 'Telnet', 25: 'SMTP', 53: 'DNS',
  80: 'HTTP', 110: 'POP3', 143: 'IMAP', 443: 'HTTPS', 445: 'SMB',
  3306: 'MySQL', 3389: 'RDP', 5432: 'PostgreSQL', 6379: 'Redis',
  8080: 'HTTP-Alt', 8443: 'HTTPS-Alt', 27017: 'MongoDB',
}

export default function PortBadges({ portOpen }) {
  if (!portOpen || typeof portOpen !== 'object') return null
  const open = Object.entries(portOpen).filter(([, v]) => v).map(([p]) => Number(p))
  if (open.length === 0) return null
  return (
    <div className="flex flex-wrap gap-1">
      {open.map(port => (
        <span key={port} className="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
          {port}{PORT_NAMES[port] ? ` ${PORT_NAMES[port]}` : ''}
        </span>
      ))}
    </div>
  )
}
