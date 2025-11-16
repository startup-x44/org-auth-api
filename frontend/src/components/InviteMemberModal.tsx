import React, { useState, useEffect } from 'react'
import { X, Mail, UserPlus, Shield, Users, Crown, Briefcase } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { organizationAPI } from '@/lib/api'
import useAuthStore from '@/store/auth'

interface InviteMemberModalProps {
  isOpen: boolean
  onClose: () => void
  onSuccess: () => void
}

interface Role {
  id: string
  name: string
  display_name: string
  description: string
  is_system: boolean
}

const InviteMemberModal: React.FC<InviteMemberModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const { organization } = useAuthStore()
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadingRoles, setLoadingRoles] = useState(false)
  const [error, setError] = useState('')
  const [roles, setRoles] = useState<Role[]>([])

  // Icon mapping for different role types
  const getRoleIcon = (roleName: string) => {
    const iconMap: Record<string, any> = {
      'student': Users,
      'issuer': Shield,
      'rto': Crown,
      'admin': Briefcase,
    }
    return iconMap[roleName.toLowerCase()] || Users
  }

  // Fetch roles when modal opens
  useEffect(() => {
    if (isOpen && organization?.organization_id) {
      fetchRoles()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, organization?.organization_id])

  const fetchRoles = async () => {
    if (!organization?.organization_id) return
    
    setLoadingRoles(true)
    setError('')
    try {
      const response = await organizationAPI.getRoles(organization.organization_id)
      if (response.success) {
        // Filter out system roles (admin) - only show non-system roles for inviting
        const availableRoles = response.data.filter((r: Role) => !r.is_system)
        setRoles(availableRoles)
        // Set first role as default if available
        if (availableRoles.length > 0 && !role) {
          setRole(availableRoles[0].name)
        }
      }
    } catch (err: any) {
      console.error('Failed to fetch roles:', err)
      setError('Failed to load roles')
    } finally {
      setLoadingRoles(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    if (!organization?.organization_id) {
      setError('Organization not found')
      setLoading(false)
      return
    }

    try {
      const response = await organizationAPI.inviteUser(organization.organization_id, {
        email,
        role: role || undefined
      })

      if (response.success) {
        setEmail('')
        setRole(roles.length > 0 ? roles[0].name : '')
        onSuccess()
        onClose()
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to send invitation')
    } finally {
      setLoading(false)
    }
  }

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
                  {loadingRoles ? (
                    <div className="flex items-center justify-center py-8">
                      <div className="animate-spin rounded-full h-8 w-8 border-2 border-blue-600 border-t-transparent" />
                    </div>
                  ) : roles.length === 0 ? (
                    <div className="text-center py-8 text-gray-500">
                      No roles available
                    </div>
                  ) : (
                    <div className="space-y-2">
                      {roles.map((roleOption) => {
                        const Icon = getRoleIcon(roleOption.name)
                        return (
                          <label
                            key={roleOption.id}
                            className={`flex items-center gap-3 p-4 rounded-xl border-2 cursor-pointer transition-all ${
                              role === roleOption.name
                                ? 'border-blue-600 bg-blue-50'
                                : 'border-gray-200 hover:border-gray-300 bg-white'
                            }`}
                          >
                            <input
                              type="radio"
                              name="role"
                              value={roleOption.name}
                              checked={role === roleOption.name}
                              onChange={(e) => setRole(e.target.value)}
                              className="w-4 h-4 text-blue-600"
                            />
                            <div className={`p-2 rounded-lg ${
                              role === roleOption.name ? 'bg-blue-100' : 'bg-gray-100'
                            }`}>
                              <Icon className={`h-5 w-5 ${
                                role === roleOption.name ? 'text-blue-600' : 'text-gray-600'
                              }`} />
                            </div>
                            <div className="flex-1">
                              <p className={`font-medium ${
                                role === roleOption.name ? 'text-blue-900' : 'text-gray-900'
                              }`}>
                                {roleOption.display_name}
                              </p>
                              <p className="text-sm text-gray-500">{roleOption.description}</p>
                            </div>
                          </label>
                        )
                      })}
                    </div>
                  )}
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
                    disabled={loading || !role || loadingRoles}
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
