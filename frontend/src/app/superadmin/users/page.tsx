'use client''use client''use client'



import { useAuth } from '../../../contexts/auth-context'

import { useRouter } from 'next/navigation'

import { useEffect, useState } from 'react'import { useAuth } from '../../../contexts/auth-context'import { useAuth } from '../../../contexts/auth-context'

import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../../components/ui/card'import { useRouter } from 'next/navigation'import { useRouter } from 'next/navigation'

import { Button } from '../../../components/ui/button'

import { Badge } from '../../../components/ui/badge'import { useEffect, useState } from 'react'import { useEffect, useState } from 'react'

import { Skeleton } from '../../../components/ui/skeleton'

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table'import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'import { SuperAdminLayout } from '../../../components/layout/superadmin-layout'

import { Input } from '../../../components/ui/input'

import { Users, Plus, Filter, MoreVertical, Edit, Trash2, Search } from 'lucide-react'import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../../../components/ui/card'import { Card } from '../../../components/ui/card'

import { useQuery } from '@tanstack/react-query'

import { Button } from '../../../components/ui/button'import { Button } from '../../../components/ui/button'

interface User {

  id: stringimport { Badge } from '../../../components/ui/badge'import { Users, Plus, Filter, MoreVertical, Edit, Trash2 } from 'lucide-react'

  email: string

  first_name: stringimport { Skeleton } from '../../../components/ui/skeleton'import { useQuery } from '@tanstack/react-query'

  last_name: string

  is_superadmin: booleanimport { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../../../components/ui/table'import { SearchInput } from '../../../components/ui/search-input'

  is_active: boolean

  created_at: stringimport { Input } from '../../../components/ui/input'import { Pagination, PaginationInfo, PaginationContainer } from '../../../components/ui';

  last_login: string | null

}import { Users, Plus, Filter, MoreVertical, Edit, Trash2, Search } from 'lucide-react'



export default function UserManagement() {import { useQuery } from '@tanstack/react-query'interface User {

  const { user, loading } = useAuth()

  const router = useRouter()  id: string

  const [searchTerm, setSearchTerm] = useState('')

interface User {  email: string

  useEffect(() => {

    if (!loading && !user) {  id: string  first_name: string

      router.push('/auth/login')

      return  email: string  last_name: string

    }

  first_name: string  is_superadmin: boolean

    if (!loading && user && !user.is_superadmin) {

      router.push('/user')  last_name: string  is_active: boolean

      return

    }  is_superadmin: boolean  created_at: string

  }, [user, loading, router])

  is_active: boolean  last_login: string | null

  const { data: users, isLoading: usersLoading } = useQuery({

    queryKey: ['users'],  created_at: string}

    queryFn: async (): Promise<User[]> => {

      await new Promise(resolve => setTimeout(resolve, 1000))  last_login: string | null

      return [

        {}export default function UserManagement() {

          id: '1',

          email: 'john.doe@example.com',  const { user, loading } = useAuth()

          first_name: 'John',

          last_name: 'Doe',export default function UserManagement() {  const router = useRouter()

          is_superadmin: false,

          is_active: true,  const { user, loading } = useAuth()  const [searchTerm, setSearchTerm] = useState('')

          created_at: '2024-01-15T10:30:00Z',

          last_login: '2024-01-20T14:30:00Z'  const router = useRouter()  const [currentPage, setCurrentPage] = useState(1)

        },

        {  const [searchTerm, setSearchTerm] = useState('')  const [itemsPerPage] = useState(10)

          id: '2',

          email: 'jane.smith@example.com',

          first_name: 'Jane',

          last_name: 'Smith',  useEffect(() => {  useEffect(() => {

          is_superadmin: false,

          is_active: true,    if (!loading && !user) {    if (!loading && !user) {

          created_at: '2024-01-10T09:15:00Z',

          last_login: '2024-01-19T16:45:00Z'      router.push('/auth/login')      router.push('/auth/login')

        },

        {      return      return

          id: '3',

          email: 'bob.wilson@example.com',    }    }

          first_name: 'Bob',

          last_name: 'Wilson',

          is_superadmin: false,

          is_active: false,    if (!loading && user && !user.is_superadmin) {    if (!loading && user && !user.is_superadmin) {

          created_at: '2024-01-05T11:20:00Z',

          last_login: null      router.push('/user')      router.push('/user')

        }

      ]      return      return

    }

  })    }    }



  if (loading) {  }, [user, loading, router])  }, [user, loading, router])

    return (

      <SuperAdminLayout>

        <div className="space-y-6">

          <Skeleton className="h-8 w-64" />  // Mock data for now - will be replaced with real API call  // Mock data for now - will be replaced with real API call

          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">

            {[1, 2, 3, 4].map((i) => (  const { data: users, isLoading: usersLoading } = useQuery({  const { data: users, isLoading: usersLoading } = useQuery({

              <Card key={i}>

                <CardContent className="pt-6">    queryKey: ['users'],    queryKey: ['users'],

                  <Skeleton className="h-20 w-full" />

                </CardContent>    queryFn: async (): Promise<User[]> => {    queryFn: async (): Promise<User[]> => {

              </Card>

            ))}      await new Promise(resolve => setTimeout(resolve, 1000))      // Mock API call

          </div>

        </div>      return [      await new Promise(resolve => setTimeout(resolve, 1000))

      </SuperAdminLayout>

    )        {      return [

  }

          id: '1',        {

  if (!user || !user.is_superadmin) {

    return null          email: 'john.doe@example.com',          id: '1',

  }

          first_name: 'John',          email: 'john.doe@example.com',

  const filteredUsers = users?.filter(u => 

    u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||          last_name: 'Doe',          first_name: 'John',

    u.first_name.toLowerCase().includes(searchTerm.toLowerCase()) ||

    u.last_name.toLowerCase().includes(searchTerm.toLowerCase())          is_superadmin: false,          last_name: 'Doe',

  )

          is_active: true,          is_superadmin: false,

  return (

    <SuperAdminLayout>          created_at: '2024-01-15T10:30:00Z',          is_active: true,

      <div className="space-y-6">

        <div className="flex justify-between items-center">          last_login: '2024-01-20T14:30:00Z'          created_at: '2024-01-15T10:30:00Z',

          <div>

            <h1 className="text-3xl font-bold">User Management</h1>        },          last_login: '2024-01-20T14:30:00Z'

            <p className="text-muted-foreground mt-2">Manage all platform users</p>

          </div>        {        },

          <Button>

            <Plus className="mr-2 h-4 w-4" />          id: '2',        {

            Add User

          </Button>          email: 'jane.smith@example.com',          id: '2',

        </div>

          first_name: 'Jane',          email: 'jane.smith@example.com',

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">

          <Card>          last_name: 'Smith',          first_name: 'Jane',

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">

              <CardTitle className="text-sm font-medium">Total Users</CardTitle>          is_superadmin: false,          last_name: 'Smith',

              <Users className="h-4 w-4 text-muted-foreground" />

            </CardHeader>          is_active: true,          is_superadmin: false,

            <CardContent>

              <div className="text-2xl font-bold">{users?.length || 0}</div>          created_at: '2024-01-10T09:15:00Z',          is_active: true,

              <p className="text-xs text-muted-foreground">All registered users</p>

            </CardContent>          last_login: '2024-01-19T16:45:00Z'          created_at: '2024-01-10T09:15:00Z',

          </Card>

        },          last_login: '2024-01-19T16:45:00Z'

          <Card>

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">        {        },

              <CardTitle className="text-sm font-medium">Active Users</CardTitle>

              <Users className="h-4 w-4 text-muted-foreground" />          id: '3',        {

            </CardHeader>

            <CardContent>          email: 'bob.wilson@example.com',          id: '3',

              <div className="text-2xl font-bold">{users?.filter(u => u.is_active).length || 0}</div>

              <p className="text-xs text-muted-foreground">Currently active</p>          first_name: 'Bob',          email: 'bob.wilson@example.com',

            </CardContent>

          </Card>          last_name: 'Wilson',          first_name: 'Bob',



          <Card>          is_superadmin: false,          last_name: 'Wilson',

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">

              <CardTitle className="text-sm font-medium">Inactive Users</CardTitle>          is_active: false,          is_superadmin: false,

              <Users className="h-4 w-4 text-muted-foreground" />

            </CardHeader>          created_at: '2024-01-05T11:20:00Z',          is_active: false,

            <CardContent>

              <div className="text-2xl font-bold">{users?.filter(u => !u.is_active).length || 0}</div>          last_login: null          created_at: '2024-01-05T11:20:00Z',

              <p className="text-xs text-muted-foreground">Disabled accounts</p>

            </CardContent>        }          last_login: null

          </Card>

      ]        }

          <Card>

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">    }      ]

              <CardTitle className="text-sm font-medium">Super Admins</CardTitle>

              <Users className="h-4 w-4 text-muted-foreground" />  })    }

            </CardHeader>

            <CardContent>  })

              <div className="text-2xl font-bold">{users?.filter(u => u.is_superadmin).length || 0}</div>

              <p className="text-xs text-muted-foreground">System administrators</p>  if (loading) {

            </CardContent>

          </Card>    return (  if (loading) {

        </div>

      <SuperAdminLayout>    return (

        <Card>

          <CardContent className="pt-6">        <div className="space-y-6">      <div className="min-h-screen flex items-center justify-center">

            <div className="flex flex-col sm:flex-row gap-4">

              <div className="relative flex-1">          <Skeleton className="h-8 w-64" />        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>

                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />

                <Input          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">      </div>

                  placeholder="Search users..."

                  value={searchTerm}            {[1, 2, 3, 4].map((i) => (    )

                  onChange={(e) => setSearchTerm(e.target.value)}

                  className="pl-8"              <Card key={i}>  }

                />

              </div>                <CardContent className="pt-6">

              <Button variant="outline">

                <Filter className="mr-2 h-4 w-4" />                  <Skeleton className="h-20 w-full" />  if (!user || !user.is_superadmin) {

                Filter

              </Button>                </CardContent>    return null

            </div>

          </CardContent>              </Card>  }

        </Card>

            ))}

        <Card>

          <CardHeader>          </div>  const filteredUsers = users?.filter(u => 

            <CardTitle>All Users</CardTitle>

            <CardDescription>        </div>    u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||

              {filteredUsers?.length || 0} users found

            </CardDescription>      </SuperAdminLayout>    u.first_name.toLowerCase().includes(searchTerm.toLowerCase()) ||

          </CardHeader>

          <CardContent>    )    u.last_name.toLowerCase().includes(searchTerm.toLowerCase())

            {usersLoading ? (

              <div className="space-y-3">  }  )

                {[1, 2, 3].map((i) => (

                  <Skeleton key={i} className="h-16 w-full" />

                ))}

              </div>  if (!user || !user.is_superadmin) {  const totalPages = Math.ceil((filteredUsers?.length || 0) / itemsPerPage)

            ) : (

              <Table>    return null  const startIndex = (currentPage - 1) * itemsPerPage

                <TableHeader>

                  <TableRow>  }  const paginatedUsers = filteredUsers?.slice(startIndex, startIndex + itemsPerPage)

                    <TableHead>User</TableHead>

                    <TableHead>Role</TableHead>

                    <TableHead>Status</TableHead>

                    <TableHead>Last Login</TableHead>  const filteredUsers = users?.filter(u =>   return (

                    <TableHead>Created</TableHead>

                    <TableHead className="text-right">Actions</TableHead>    u.email.toLowerCase().includes(searchTerm.toLowerCase()) ||    <SuperAdminLayout>

                  </TableRow>

                </TableHeader>    u.first_name.toLowerCase().includes(searchTerm.toLowerCase()) ||      <div className="space-y-6">

                <TableBody>

                  {filteredUsers?.map((userData) => (    u.last_name.toLowerCase().includes(searchTerm.toLowerCase())        {/* Header */}

                    <TableRow key={userData.id}>

                      <TableCell>  )        <div className="flex justify-between items-center">

                        <div className="flex items-center gap-3">

                          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">          <div>

                            <span className="text-sm font-medium">

                              {userData.first_name[0]}{userData.last_name[0]}  return (            <h1 className="text-3xl font-bold text-gray-900">User Management</h1>

                            </span>

                          </div>    <SuperAdminLayout>            <p className="text-gray-600 mt-2">Manage all platform users</p>

                          <div>

                            <div className="font-medium">      <div className="space-y-6">          </div>

                              {userData.first_name} {userData.last_name}

                            </div>        {/* Header */}          <Button className="flex items-center gap-2">

                            <div className="text-sm text-muted-foreground">

                              {userData.email}        <div className="flex justify-between items-center">            <Plus className="h-4 w-4" />

                            </div>

                          </div>          <div>            Add User

                        </div>

                      </TableCell>            <h1 className="text-3xl font-bold">User Management</h1>          </Button>

                      <TableCell>

                        <Badge variant={userData.is_superadmin ? "default" : "secondary"}>            <p className="text-muted-foreground mt-2">Manage all platform users</p>        </div>

                          {userData.is_superadmin ? 'Super Admin' : 'User'}

                        </Badge>          </div>

                      </TableCell>

                      <TableCell>          <Button>        {/* Stats Cards */}

                        <Badge variant={userData.is_active ? "default" : "destructive"}>

                          {userData.is_active ? 'Active' : 'Inactive'}            <Plus className="mr-2 h-4 w-4" />        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">

                        </Badge>

                      </TableCell>            Add User          <Card className="p-6">

                      <TableCell>

                        {userData.last_login           </Button>            <div className="flex items-center">

                          ? new Date(userData.last_login).toLocaleDateString()

                          : 'Never'        </div>              <div className="p-2 bg-blue-100 rounded-lg">

                        }

                      </TableCell>                <Users className="h-6 w-6 text-blue-600" />

                      <TableCell>

                        {new Date(userData.created_at).toLocaleDateString()}        {/* Stats Cards */}              </div>

                      </TableCell>

                      <TableCell className="text-right">        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">              <div className="ml-4">

                        <div className="flex items-center justify-end gap-2">

                          <Button variant="ghost" size="icon">          <Card>                <p className="text-sm font-medium text-gray-600">Total Users</p>

                            <Edit className="h-4 w-4" />

                          </Button>            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">                <p className="text-2xl font-bold text-gray-900">{users?.length || 0}</p>

                          <Button variant="ghost" size="icon">

                            <Trash2 className="h-4 w-4" />              <CardTitle className="text-sm font-medium">Total Users</CardTitle>              </div>

                          </Button>

                          <Button variant="ghost" size="icon">              <Users className="h-4 w-4 text-muted-foreground" />            </div>

                            <MoreVertical className="h-4 w-4" />

                          </Button>            </CardHeader>          </Card>

                        </div>

                      </TableCell>            <CardContent>

                    </TableRow>

                  ))}              <div className="text-2xl font-bold">{users?.length || 0}</div>          <Card className="p-6">

                </TableBody>

              </Table>              <p className="text-xs text-muted-foreground">All registered users</p>            <div className="flex items-center">

            )}

          </CardContent>            </CardContent>              <div className="p-2 bg-green-100 rounded-lg">

        </Card>

      </div>          </Card>                <Users className="h-6 w-6 text-green-600" />

    </SuperAdminLayout>

  )              </div>

}

          <Card>              <div className="ml-4">

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">                <p className="text-sm font-medium text-gray-600">Active Users</p>

              <CardTitle className="text-sm font-medium">Active Users</CardTitle>                <p className="text-2xl font-bold text-gray-900">

              <Users className="h-4 w-4 text-muted-foreground" />                  {users?.filter(u => u.is_active).length || 0}

            </CardHeader>                </p>

            <CardContent>              </div>

              <div className="text-2xl font-bold">{users?.filter(u => u.is_active).length || 0}</div>            </div>

              <p className="text-xs text-muted-foreground">Currently active</p>          </Card>

            </CardContent>

          </Card>          <Card className="p-6">

            <div className="flex items-center">

          <Card>              <div className="p-2 bg-red-100 rounded-lg">

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">                <Users className="h-6 w-6 text-red-600" />

              <CardTitle className="text-sm font-medium">Inactive Users</CardTitle>              </div>

              <Users className="h-4 w-4 text-muted-foreground" />              <div className="ml-4">

            </CardHeader>                <p className="text-sm font-medium text-gray-600">Inactive Users</p>

            <CardContent>                <p className="text-2xl font-bold text-gray-900">

              <div className="text-2xl font-bold">{users?.filter(u => !u.is_active).length || 0}</div>                  {users?.filter(u => !u.is_active).length || 0}

              <p className="text-xs text-muted-foreground">Disabled accounts</p>                </p>

            </CardContent>              </div>

          </Card>            </div>

          </Card>

          <Card>

            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">          <Card className="p-6">

              <CardTitle className="text-sm font-medium">Super Admins</CardTitle>            <div className="flex items-center">

              <Users className="h-4 w-4 text-muted-foreground" />              <div className="p-2 bg-purple-100 rounded-lg">

            </CardHeader>                <Users className="h-6 w-6 text-purple-600" />

            <CardContent>              </div>

              <div className="text-2xl font-bold">{users?.filter(u => u.is_superadmin).length || 0}</div>              <div className="ml-4">

              <p className="text-xs text-muted-foreground">System administrators</p>                <p className="text-sm font-medium text-gray-600">Super Admins</p>

            </CardContent>                <p className="text-2xl font-bold text-gray-900">

          </Card>                  {users?.filter(u => u.is_superadmin).length || 0}

        </div>                </p>

              </div>

        {/* Search and Filters */}            </div>

        <Card>          </Card>

          <CardContent className="pt-6">        </div>

            <div className="flex flex-col sm:flex-row gap-4">

              <div className="relative flex-1">        {/* Search and Filters */}

                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />        <Card className="p-6">

                <Input          <div className="flex flex-col sm:flex-row gap-4">

                  placeholder="Search users..."            <div className="flex-1">

                  value={searchTerm}              <SearchInput

                  onChange={(e) => setSearchTerm(e.target.value)}                placeholder="Search users..."

                  className="pl-8"                onSearch={setSearchTerm}

                />                loading={usersLoading}

              </div>              />

              <Button variant="outline">            </div>

                <Filter className="mr-2 h-4 w-4" />            <Button variant="outline" className="flex items-center gap-2">

                Filter              <Filter className="h-4 w-4" />

              </Button>              Filter

            </div>            </Button>

          </CardContent>          </div>

        </Card>        </Card>



        {/* Users Table */}        {/* Users Table */}

        <Card>        <PaginationContainer

          <CardHeader>          pagination={

            <CardTitle>All Users</CardTitle>            <Pagination

            <CardDescription>              currentPage={currentPage}

              {filteredUsers?.length || 0} users found              totalPages={totalPages}

            </CardDescription>              onPageChange={setCurrentPage}

          </CardHeader>            />

          <CardContent>          }

            {usersLoading ? (          info={

              <div className="space-y-3">            <PaginationInfo

                {[1, 2, 3].map((i) => (              currentPage={currentPage}

                  <Skeleton key={i} className="h-16 w-full" />              totalItems={filteredUsers?.length || 0}

                ))}              itemsPerPage={itemsPerPage}

              </div>            />

            ) : (          }

              <Table>        >

                <TableHeader>          <Card className="p-6">

                  <TableRow>            <h3 className="text-lg font-semibold text-gray-900 mb-4">All Users</h3>

                    <TableHead>User</TableHead>            

                    <TableHead>Role</TableHead>            {usersLoading ? (

                    <TableHead>Status</TableHead>              <div className="flex items-center justify-center py-8">

                    <TableHead>Last Login</TableHead>                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>

                    <TableHead>Created</TableHead>              </div>

                    <TableHead className="text-right">Actions</TableHead>            ) : (

                  </TableRow>              <div className="overflow-x-auto">

                </TableHeader>                <table className="min-w-full divide-y divide-gray-200">

                <TableBody>                  <thead className="bg-gray-50">

                  {filteredUsers?.map((userData) => (                    <tr>

                    <TableRow key={userData.id}>                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">

                      <TableCell>                        User

                        <div className="flex items-center gap-3">                      </th>

                          <div className="flex h-10 w-10 items-center justify-center rounded-full bg-muted">                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">

                            <span className="text-sm font-medium">                        Role

                              {userData.first_name[0]}{userData.last_name[0]}                      </th>

                            </span>                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">

                          </div>                        Status

                          <div>                      </th>

                            <div className="font-medium">                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">

                              {userData.first_name} {userData.last_name}                        Last Login

                            </div>                      </th>

                            <div className="text-sm text-muted-foreground">                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">

                              {userData.email}                        Created

                            </div>                      </th>

                          </div>                      <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">

                        </div>                        Actions

                      </TableCell>                      </th>

                      <TableCell>                    </tr>

                        <Badge variant={userData.is_superadmin ? "default" : "secondary"}>                  </thead>

                          {userData.is_superadmin ? 'Super Admin' : 'User'}                  <tbody className="bg-white divide-y divide-gray-200">

                        </Badge>                    {paginatedUsers?.map((user) => (

                      </TableCell>                    <tr key={user.id} className="hover:bg-gray-50">

                      <TableCell>                      <td className="px-6 py-4 whitespace-nowrap">

                        <Badge variant={userData.is_active ? "default" : "destructive"}>                        <div className="flex items-center">

                          {userData.is_active ? 'Active' : 'Inactive'}                          <div className="flex-shrink-0 h-10 w-10">

                        </Badge>                            <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">

                      </TableCell>                              <span className="text-sm font-medium text-gray-700">

                      <TableCell>                                {user.first_name[0]}{user.last_name[0]}

                        {userData.last_login                               </span>

                          ? new Date(userData.last_login).toLocaleDateString()                            </div>

                          : 'Never'                          </div>

                        }                          <div className="ml-4">

                      </TableCell>                            <div className="text-sm font-medium text-gray-900">

                      <TableCell>                              {user.first_name} {user.last_name}

                        {new Date(userData.created_at).toLocaleDateString()}                            </div>

                      </TableCell>                            <div className="text-sm text-gray-500">

                      <TableCell className="text-right">                              {user.email}

                        <div className="flex items-center justify-end gap-2">                            </div>

                          <Button variant="ghost" size="icon">                          </div>

                            <Edit className="h-4 w-4" />                        </div>

                          </Button>                      </td>

                          <Button variant="ghost" size="icon">                      <td className="px-6 py-4 whitespace-nowrap">

                            <Trash2 className="h-4 w-4" />                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${

                          </Button>                          user.is_superadmin 

                          <Button variant="ghost" size="icon">                            ? 'bg-purple-100 text-purple-800' 

                            <MoreVertical className="h-4 w-4" />                            : 'bg-gray-100 text-gray-800'

                          </Button>                        }`}>

                        </div>                          {user.is_superadmin ? 'Super Admin' : 'User'}

                      </TableCell>                        </span>

                    </TableRow>                      </td>

                  ))}                      <td className="px-6 py-4 whitespace-nowrap">

                </TableBody>                        <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${

              </Table>                          user.is_active 

            )}                            ? 'bg-green-100 text-green-800' 

          </CardContent>                            : 'bg-red-100 text-red-800'

        </Card>                        }`}>

      </div>                          {user.is_active ? 'Active' : 'Inactive'}

    </SuperAdminLayout>                        </span>

  )                      </td>

}                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">

                        {user.last_login 
                          ? new Date(user.last_login).toLocaleDateString()
                          : 'Never'
                        }
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {new Date(user.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                        <div className="flex items-center justify-end space-x-2">
                          <button className="text-blue-600 hover:text-blue-900">
                            <Edit className="h-4 w-4" />
                          </button>
                          <button className="text-red-600 hover:text-red-900">
                            <Trash2 className="h-4 w-4" />
                          </button>
                          <button className="text-gray-600 hover:text-gray-900">
                            <MoreVertical className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </PaginationContainer>
      </div>
    </SuperAdminLayout>
  )
}