import { useTranslation } from 'react-i18next'

export default function PermissionDenied({ message }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center text-gray-500 dark:text-gray-400">
      <svg className="w-12 h-12 mb-4 text-gray-300 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
      </svg>
      <p className="text-sm font-medium">{message ?? t('permissionDenied.defaultMessage')}</p>
    </div>
  )
}
