// Two-option yes/no choice used by wizard steps. `value` is 'true' | 'false' |
// undefined (not chosen yet); `onChange` receives 'true' or 'false'.
export default function ChoiceButtons({ value, onChange, yesLabel, noLabel }) {
  const base = 'w-full py-3 rounded-lg font-medium border transition'
  const selected = 'bg-blue-600 text-white border-blue-600'
  const idle = 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50'
  return (
    <div role="radiogroup" className="flex flex-col gap-3">
      <button
        type="button"
        role="radio"
        aria-checked={value === 'true'}
        onClick={() => onChange('true')}
        className={`${base} ${value === 'true' ? selected : idle}`}
      >
        {yesLabel}
      </button>
      <button
        type="button"
        role="radio"
        aria-checked={value === 'false'}
        onClick={() => onChange('false')}
        className={`${base} ${value === 'false' ? selected : idle}`}
      >
        {noLabel}
      </button>
    </div>
  )
}
