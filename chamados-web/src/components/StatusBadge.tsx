import { STATUS } from '../constants'
import type { Status } from '../types'

export default function StatusBadge({ status }: { status: Status }) {
  return <span className={`badge badge-${status}`}>{STATUS[status]}</span>
}
