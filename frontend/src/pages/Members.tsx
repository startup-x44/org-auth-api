import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { 
  ArrowLeft, 
  Users, 
  UserPlus, 
  Crown, 
  Shield, 
  MoreVertical,
  Mail,
  Calendar,
  Search,
  Clock,
  RotateCw,
  X
} from 'lucide-react';
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import LoadingSpinner from '@/components/ui/loading-spinner'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useToast } from '@/hooks/use-toast'
import useAuthStore from '@/store/auth'
import InviteMemberModal from '@/components/InviteMemberModal'
import api from '@/lib/axios-instance'

interface Member {
  user_id: string
  email: string
  first_name: string
  last_name: string
  role: string
  status: string
  joined_at: string
}

interface Invitation {
  id: string
  email: string
  role: string
  status: string
  invited_by: string
  created_at: string
  expires_at: string
}

const Members = () => {
  const navigate = useNavigate()
  const { organization, user } = useAuthStore()
  const { toast } = useToast()
  const [searchQuery, setSearchQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false)
  const [members, setMembers] = useState<Member[]>([])
  const [invitations, setInvitations] = useState<Invitation[]>([])
  const [cancelInvitationId, setCancelInvitationId] = useState<string | null>(null)
  const [stats, setStats] = useState({
    total: 0,
    admins: 0,
    pending: 0
  })

  useEffect(() => {
    if (organization?.organization_id) {
      fetchMembers()
      fetchInvitations()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [organization?.organization_id])

  const fetchMembers = async () => {
    try {
      setLoading(true)
      const response = await api.get(`/organizations/${organization?.organization_id}/members`)
      
      if (response.data.success && response.data.data) {
        setMembers(response.data.data)
        calculateStats(response.data.data)
      }
    } catch (err: any) {
      // 401 errors are handled by axios interceptor with redirect
      if (err.response?.status === 401) {
        return // Let interceptor handle the redirect
      }
      console.error('Failed to fetch members:', err)
    } finally {
      setLoading(false)
    }
  }

  const fetchInvitations = async () => {
    try {
      const response = await api.get(`/organizations/${organization?.organization_id}/invitations`)
      
      if (response.data.success && response.data.data) {
        setInvitations(response.data.data)
        setStats(prev => ({ ...prev, pending: response.data.data.length }))
      }
    } catch (err: any) {
      // 401 errors are handled by axios interceptor with redirect
      if (err.response?.status === 401) {
        return // Let interceptor handle the redirect
      }
      console.error('Failed to fetch invitations:', err)
    }
  }

  const calculateStats = (membersList: Member[]) => {
    const admins = membersList.filter(m => 
      m.role?.toLowerCase() === 'rto' || m.role?.toLowerCase() === 'issuer' || m.role?.toLowerCase() === 'owner' || m.role?.toLowerCase() === 'admin'
    ).length
    
    setStats(prev => ({
      ...prev,
      total: membersList.length,
      admins: admins
    }))
  }

  const handleInviteSuccess = () => {
    fetchInvitations()
  }

  const handleCancelInvitation = async (invitationId: string) => {
    setCancelInvitationId(invitationId)
  }

  const confirmCancelInvitation = async () => {
    if (!cancelInvitationId) return

    try {
      const response = await api.delete(`/organizations/${organization?.organization_id}/invitations/${cancelInvitationId}`)
      
      if (response.data.success) {
        fetchInvitations()
        toast({
          title: "Invitation cancelled",
          description: "The invitation has been cancelled successfully.",
        })
      }
    } catch (err: any) {
      console.error('Failed to cancel invitation:', err)
      toast({
        title: "Error",
        description: err.response?.data?.message || 'Failed to cancel invitation',
        variant: "destructive",
      })
    } finally {
      setCancelInvitationId(null)
    }
  }

  const handleResendInvitation = async (invitationId: string) => {
    try {
      const response = await api.post(`/organizations/${organization?.organization_id}/invitations/${invitationId}/resend`)
      
      if (response.data.success) {
        toast({
          title: "Invitation resent",
          description: "The invitation has been resent successfully.",
        })
      }
    } catch (err: any) {
      console.error('Failed to resend invitation:', err)
      toast({
        title: "Error",
        description: err.response?.data?.message || 'Failed to resend invitation',
        variant: "destructive",
      })
    }
  }

  const getRoleBadge = (role: string) => {
    const roleConfig = {
      rto: { color: 'bg-amber-100 text-amber-800 border-amber-200', icon: Crown },
      issuer: { color: 'bg-blue-100 text-blue-800 border-blue-200', icon: Shield },
      student: { color: 'bg-green-100 text-green-800 border-green-200', icon: Users },
      owner: { color: 'bg-purple-100 text-purple-800 border-purple-200', icon: Crown },
      admin: { color: 'bg-indigo-100 text-indigo-800 border-indigo-200', icon: Shield },
      member: { color: 'bg-gray-100 text-gray-800 border-gray-200', icon: Users },
    }

    const config = roleConfig[role?.toLowerCase() as keyof typeof roleConfig] || roleConfig.student
    const Icon = config.icon

    return (
      <Badge className={`${config.color} flex items-center gap-1.5 px-3 py-1 border`}>
        <Icon className="h-3.5 w-3.5" />
        <span className="capitalize font-medium">{role}</span>
      </Badge>
    )
  }

  const getInitials = (firstName?: string, lastName?: string, email?: string) => {
    if (firstName && lastName) {
      return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase()
    }
    if (email) {
      return email.substring(0, 2).toUpperCase()
    }
    return 'U'
  }

  const handleBack = () => {
    navigate('/dashboard')
  }

  const filteredMembers = members.filter(member => 
    member.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
    member.first_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    member.last_name?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const filteredInvitations = invitations.filter(invitation =>
    invitation.email.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50/30 to-purple-50/20">
      {/* Floating orbs background */}
      <div className="fixed inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-0 -left-4 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob" />
        <div className="absolute top-0 -right-4 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000" />
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000" />
      </div>

      {/* Header */}
      <header className="relative bg-white/80 backdrop-blur-xl border-b border-gray-200/50 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleBack}
              className="flex items-center gap-2"
            >
              <ArrowLeft className="h-4 w-4" />
              <span>Back</span>
            </Button>
            <h1 className="text-xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
              Team Members
            </h1>
            <Button
              size="sm"
              onClick={() => setIsInviteModalOpen(true)}
              className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
            >
              <UserPlus className="h-4 w-4 mr-2" />
              Invite Member
            </Button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="space-y-6"
        >
          {/* Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <Card className="border-0 shadow-lg bg-gradient-to-br from-blue-50 to-blue-100/50">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-blue-600">Total Members</p>
                    <p className="text-3xl font-bold text-blue-900 mt-1">{stats.total}</p>
                  </div>
                  <div className="p-3 bg-blue-600 rounded-xl">
                    <Users className="h-6 w-6 text-white" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg bg-gradient-to-br from-purple-50 to-purple-100/50">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-purple-600">Issuers & RTOs</p>
                    <p className="text-3xl font-bold text-purple-900 mt-1">{stats.admins}</p>
                  </div>
                  <div className="p-3 bg-purple-600 rounded-xl">
                    <Shield className="h-6 w-6 text-white" />
                  </div>
                </div>
              </CardContent>
            </Card>

            <Card className="border-0 shadow-lg bg-gradient-to-br from-amber-50 to-amber-100/50">
              <CardContent className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-amber-600">Pending Invites</p>
                    <p className="text-3xl font-bold text-amber-900 mt-1">{stats.pending}</p>
                  </div>
                  <div className="p-3 bg-amber-600 rounded-xl">
                    <Mail className="h-6 w-6 text-white" />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Members List */}
          <Card className="border-0 shadow-lg bg-white/80 backdrop-blur">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-gray-900">Members of {organization?.organization_name}</CardTitle>
                  <CardDescription>
                    Manage your organization's team members and roles
                  </CardDescription>
                </div>
              </div>
              <div className="pt-4">
                <div className="relative">
                  <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search members by name or email..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex justify-center py-12">
                  <LoadingSpinner size="lg" />
                </div>
              ) : filteredMembers.length === 0 && filteredInvitations.length === 0 ? (
                <div className="text-center py-12">
                  <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600">No members or invitations found</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {/* Active Members */}
                  {filteredMembers.map((member) => (
                    <motion.div
                      key={member.user_id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      className="flex items-center justify-between p-4 bg-gradient-to-r from-gray-50 to-transparent rounded-xl border border-gray-100 hover:border-blue-200 transition-all"
                    >
                      <div className="flex items-center gap-4">
                        <Avatar className="h-12 w-12 ring-2 ring-purple-100">
                          <AvatarFallback className="bg-gradient-to-br from-blue-500 to-purple-500 text-white font-semibold">
                            {getInitials(member.first_name, member.last_name, member.email)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="flex items-center gap-2">
                            <p className="font-semibold text-gray-900">
                              {member.first_name} {member.last_name}
                              {member.email === user?.email && (
                                <span className="text-sm text-gray-500 font-normal ml-2">(You)</span>
                              )}
                            </p>
                          </div>
                          <p className="text-sm text-gray-600 flex items-center gap-1">
                            <Mail className="h-3 w-3" />
                            {member.email}
                          </p>
                          <p className="text-xs text-gray-500 flex items-center gap-1 mt-1">
                            <Calendar className="h-3 w-3" />
                            Joined {new Date(member.joined_at).toLocaleDateString()}
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        {getRoleBadge(member.role)}
                        <Badge 
                          className={`${
                            member.status === 'active' 
                              ? 'bg-green-100 text-green-800' 
                              : 'bg-gray-100 text-gray-800'
                          }`}
                        >
                          {member.status}
                        </Badge>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </div>
                    </motion.div>
                  ))}

                  {/* Pending Invitations */}
                  {filteredInvitations.length > 0 && (
                    <>
                      <div className="pt-6 mt-6 border-t border-gray-200">
                        <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
                          <UserPlus className="h-5 w-5 text-purple-600" />
                          Pending Invitations
                        </h3>
                      </div>
                      {filteredInvitations.map((invitation) => (
                        <motion.div
                          key={invitation.id}
                          initial={{ opacity: 0, x: -20 }}
                          animate={{ opacity: 1, x: 0 }}
                          className="flex items-center justify-between p-4 bg-gradient-to-r from-purple-50 to-transparent rounded-xl border border-purple-100 hover:border-purple-200 transition-all"
                        >
                          <div className="flex items-center gap-4">
                            <div className="h-12 w-12 rounded-full bg-gradient-to-br from-purple-400 to-pink-400 flex items-center justify-center">
                              <Mail className="h-6 w-6 text-white" />
                            </div>
                            <div>
                              <div className="flex items-center gap-2">
                                <p className="font-semibold text-gray-900">{invitation.email}</p>
                                <Badge className="bg-yellow-100 text-yellow-800">
                                  Pending
                                </Badge>
                              </div>
                              <p className="text-xs text-gray-500 flex items-center gap-1 mt-1">
                                <Calendar className="h-3 w-3" />
                                Invited {new Date(invitation.created_at).toLocaleDateString()}
                              </p>
                              <p className="text-xs text-gray-500 flex items-center gap-1 mt-1">
                                <Clock className="h-3 w-3" />
                                Expires {new Date(invitation.expires_at).toLocaleDateString()}
                              </p>
                            </div>
                          </div>

                          <div className="flex items-center gap-3">
                            {getRoleBadge(invitation.role)}
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleResendInvitation(invitation.id)}
                              className="border-blue-200 text-blue-600 hover:bg-blue-50"
                            >
                              <RotateCw className="h-4 w-4 mr-1" />
                              Resend
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleCancelInvitation(invitation.id)}
                              className="border-red-200 text-red-600 hover:bg-red-50"
                            >
                              <X className="h-4 w-4 mr-1" />
                              Cancel
                            </Button>
                          </div>
                        </motion.div>
                      ))}
                    </>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </motion.div>
      </main>

      {/* Invite Modal */}
      <InviteMemberModal
        isOpen={isInviteModalOpen}
        onClose={() => setIsInviteModalOpen(false)}
        onSuccess={handleInviteSuccess}
      />

      {/* Cancel Invitation Confirmation Dialog */}
      <AlertDialog open={!!cancelInvitationId} onOpenChange={() => setCancelInvitationId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Cancel Invitation?</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to cancel this invitation? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>No, keep it</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmCancelInvitation}
              className="bg-red-600 hover:bg-red-700"
            >
              Yes, cancel invitation
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

export default Members
