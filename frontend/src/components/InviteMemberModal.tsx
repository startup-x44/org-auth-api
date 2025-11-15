import React, { useState } from 'react'
import { X, Mail, UserPlus, Shield, Users, Crown } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import api from '@/lib/axios-instance'
import useAuthStore from '@/store/auth'

interface InviteMemberModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
}

const InviteMemberModal: React.FC<InviteMemberModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const { organization } = useAuthStore()
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('student')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const response = await api.post(`/organizations/${organization?.organization_id}/members`, {
        email,
        role
      })

      if (response.data.success) {
        setEmail('')
        setRole('student')
        onSuccess()
        onClose()
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to send invitation')
    } finally {
      setLoading(false)
    }
  }

  const roleOptions = [
    { value: 'student', label: 'Student', icon: Users, description: 'Can view and access credentials' },
    { value: 'issuer', label: 'Issuer', icon: Shield, description: 'Can issue and verify credentials' },
    { value: 'rto', label: 'RTO', icon: Crown, description: 'Registered Training Organization' }
  ]

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          {/* Backdrop */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
            className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
          />

          {/* Modal */}
          <div className="fixed inset-0 flex items-center justify-center z-50 p-4">
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 20 }}
              transition={{ duration: 0.2 }}
              className="bg-white rounded-2xl shadow-2xl max-w-md w-full overflow-hidden"
            >
              {/* Header */}
              <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-6 text-white">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-white/20 rounded-lg">
                      <UserPlus className="h-6 w-6" />
                    </div>
                    <div>
                      <h2 className="text-xl font-bold">Invite Team Member</h2>
                      <p className="text-blue-100 text-sm">Send an invitation to join your organization</p>
                    </div>
                  </div>
                  <button
                    onClick={onClose}
                    className="p-1 hover:bg-white/20 rounded-lg transition-colors"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>
              </div>

              {/* Form */}
              <form onSubmit={handleSubmit} className="p-6 space-y-6">
                {error && (
                  <div className="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg text-sm">
                    {error}
                  </div>
                )}

                {/* Email Input */}
                <div className="space-y-2">
                  <Label htmlFor="email" className="text-sm font-medium text-gray-700">
                    Email Address
                  </Label>
                  <div className="relative">
                    <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
                    <Input
                      id="email"
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="colleague@example.com"
                      className="pl-10 h-12"
                      required
                    />
                  </div>
                </div>

                {/* Role Selection */}
                <div className="space-y-2">
                  <Label className="text-sm font-medium text-gray-700">
                    Role
                  </Label>
                  <div className="space-y-2">
                    {roleOptions.map((option) => {
                      const Icon = option.icon
                      return (
                        <label
                          key={option.value}
                          className={`flex items-center gap-3 p-4 rounded-xl border-2 cursor-pointer transition-all ${
                            role === option.value
                              ? 'border-blue-600 bg-blue-50'
                              : 'border-gray-200 hover:border-gray-300 bg-white'
                          }`}
                        >
                          <input
                            type="radio"
                            name="role"
                            value={option.value}
                            checked={role === option.value}
                            onChange={(e) => setRole(e.target.value)}
                            className="w-4 h-4 text-blue-600"
                          />
                          <div className={`p-2 rounded-lg ${
                            role === option.value ? 'bg-blue-100' : 'bg-gray-100'
                          }`}>
                            <Icon className={`h-5 w-5 ${
                              role === option.value ? 'text-blue-600' : 'text-gray-600'
                            }`} />
                          </div>
                          <div className="flex-1">
                            <p className={`font-medium ${
                              role === option.value ? 'text-blue-900' : 'text-gray-900'
                            }`}>
                              {option.label}
                            </p>
                            <p className="text-sm text-gray-500">{option.description}</p>
                          </div>
                        </label>
                      )
                    })}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex gap-3 pt-4">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={onClose}
                    className="flex-1"
                    disabled={loading}
                  >
                    Cancel
                  </Button>
                  <Button
                    type="submit"
                    className="flex-1 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                    disabled={loading}
                  >
                    {loading ? (
                      <>
                        <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent mr-2" />
                        Sending...
                      </>
                    ) : (
                      <>
                        <Mail className="h-4 w-4 mr-2" />
                        Send Invitation
                      </>
                    )}
                  </Button>
                </div>
              </form>
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  )
}

export default InviteMemberModal
