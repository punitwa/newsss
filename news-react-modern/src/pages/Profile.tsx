import { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Badge } from '@/components/ui/badge';
import { 
  User, 
  Mail, 
  Calendar, 
  Settings, 
  Bell, 
  Shield, 
  Edit3,
  Save,
  X,
  BookOpen,
  Heart,
  TrendingUp,
  Camera,
  Globe,
  Award,
  Activity,
  Clock
} from 'lucide-react';
import { format } from 'date-fns';

const Profile = () => {
  const { authState } = useAuth();
  const navigate = useNavigate();
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    first_name: authState.user?.first_name || '',
    last_name: authState.user?.last_name || '',
    username: authState.user?.username || '',
  });

  const handleSave = () => {
    // TODO: Implement profile update API call
    console.log('Saving profile:', editForm);
    setIsEditing(false);
  };

  const handleCancel = () => {
    setEditForm({
      first_name: authState.user?.first_name || '',
      last_name: authState.user?.last_name || '',
      username: authState.user?.username || '',
    });
    setIsEditing(false);
  };

  if (!authState.isAuthenticated || !authState.user) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50 flex items-center justify-center">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6">
            <div className="text-center">
              <Shield className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <h3 className="text-lg font-semibold">Access Denied</h3>
              <p className="text-muted-foreground">Please log in to view your profile.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  const { user } = authState;

  // Get user initials for avatar
  const getInitials = (firstName: string, lastName: string) => {
    return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-blue-50/30">
      <div className="container mx-auto px-4 py-8">
        <div className="max-w-6xl mx-auto">
          {/* Modern Header with Profile Card */}
          <div className="relative mb-12">
            {/* Background Pattern */}
            <div className="absolute inset-0 bg-gradient-to-r from-blue-600/10 via-purple-600/10 to-cyan-600/10 rounded-3xl"></div>
            
            {/* Header Content */}
            <div className="relative bg-white/80 backdrop-blur-sm rounded-3xl border border-white/20 shadow-xl p-8">
              <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-8">
                {/* Profile Info */}
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-6">
                  {/* Avatar */}
                  <div className="relative group">
                    {user.avatar ? (
                      <img
                        src={user.avatar}
                        alt={`${user.first_name} ${user.last_name}`}
                        className="h-24 w-24 rounded-2xl object-cover ring-4 ring-white shadow-lg"
                      />
                    ) : (
                      <div className="h-24 w-24 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center ring-4 ring-white shadow-lg">
                        <span className="text-white text-2xl font-bold">
                          {getInitials(user.first_name, user.last_name)}
                        </span>
                      </div>
                    )}
                    <button className="absolute -bottom-2 -right-2 p-2 bg-white rounded-full shadow-lg hover:shadow-xl transition-all duration-200 border border-gray-200 hover:border-blue-300 group-hover:scale-110">
                      <Camera className="h-4 w-4 text-gray-600" />
                    </button>
                  </div>

                  {/* User Details */}
                  <div className="space-y-2">
                    <div className="flex items-center gap-3">
                      <h1 className="text-3xl font-bold bg-gradient-to-r from-gray-900 to-gray-700 bg-clip-text text-transparent">
                        {user.first_name} {user.last_name}
                      </h1>
                      {user.is_admin && (
                        <Badge className="bg-gradient-to-r from-purple-500 to-purple-600 text-white border-0">
                          <Award className="h-3 w-3 mr-1" />
                          Admin
                        </Badge>
                      )}
                    </div>
                    <p className="text-lg text-gray-600 font-medium">@{user.username}</p>
                    <div className="flex items-center gap-4 text-sm text-gray-500">
                      <div className="flex items-center gap-1">
                        <Mail className="h-4 w-4" />
                        {user.email}
                      </div>
                      <div className="flex items-center gap-1">
                        <Calendar className="h-4 w-4" />
                        Member since {format(new Date(user.created_at), 'MMM yyyy')}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Status and Actions */}
                <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4">
                  <div className="flex items-center gap-2">
                    <Badge 
                      variant={user.is_active ? "default" : "secondary"}
                      className={`${user.is_active 
                        ? "bg-green-100 text-green-800 border-green-200" 
                        : "bg-gray-100 text-gray-600 border-gray-200"
                      } px-3 py-1`}
                    >
                      <div className={`h-2 w-2 rounded-full mr-2 ${user.is_active ? "bg-green-500" : "bg-gray-400"}`}></div>
                      {user.is_active ? "Active" : "Inactive"}
                    </Badge>
                  </div>
                  <Button 
                    variant="outline" 
                    className="bg-white/50 border-white/20 hover:bg-white/80 backdrop-blur-sm"
                    onClick={() => setIsEditing(!isEditing)}
                  >
                    <Edit3 className="h-4 w-4 mr-2" />
                    {isEditing ? 'Cancel Edit' : 'Edit Profile'}
                  </Button>
                </div>
              </div>
            </div>
          </div>

          <div className="grid gap-8 lg:grid-cols-3">
            {/* Main Content Area */}
            <div className="lg:col-span-2 space-y-8">
              {/* Profile Information Card */}
              {isEditing ? (
                <Card className="border-0 shadow-xl bg-white/70 backdrop-blur-sm">
                  <CardHeader className="pb-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-xl font-semibold text-gray-900">Edit Profile</CardTitle>
                        <CardDescription className="text-gray-600">
                          Update your personal information
                        </CardDescription>
                      </div>
                      <div className="flex space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleCancel}
                          className="hover:bg-gray-50"
                        >
                          <X className="h-4 w-4 mr-2" />
                          Cancel
                        </Button>
                        <Button
                          size="sm"
                          onClick={handleSave}
                          className="bg-blue-600 hover:bg-blue-700"
                        >
                          <Save className="h-4 w-4 mr-2" />
                          Save Changes
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-6">
                    <div className="grid gap-6 md:grid-cols-2">
                      <div className="space-y-3">
                        <Label htmlFor="first_name" className="text-sm font-medium text-gray-700">First Name</Label>
                        <Input
                          id="first_name"
                          value={editForm.first_name}
                          onChange={(e) => setEditForm({ ...editForm, first_name: e.target.value })}
                          className="border-gray-200 focus:border-blue-500 focus:ring-blue-500/20"
                        />
                      </div>
                      <div className="space-y-3">
                        <Label htmlFor="last_name" className="text-sm font-medium text-gray-700">Last Name</Label>
                        <Input
                          id="last_name"
                          value={editForm.last_name}
                          onChange={(e) => setEditForm({ ...editForm, last_name: e.target.value })}
                          className="border-gray-200 focus:border-blue-500 focus:ring-blue-500/20"
                        />
                      </div>
                    </div>
                    
                    <div className="space-y-3">
                      <Label htmlFor="username" className="text-sm font-medium text-gray-700">Username</Label>
                      <Input
                        id="username"
                        value={editForm.username}
                        onChange={(e) => setEditForm({ ...editForm, username: e.target.value })}
                        className="border-gray-200 focus:border-blue-500 focus:ring-blue-500/20"
                      />
                    </div>
                  </CardContent>
                </Card>
              ) : (
                <div className="grid gap-6 md:grid-cols-2">
                  {/* Personal Info */}
                  <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-gray-50/50 hover:shadow-xl transition-all duration-300">
                    <CardContent className="p-6">
                      <div className="flex items-center gap-3 mb-4">
                        <div className="p-2 bg-blue-100 rounded-lg">
                          <User className="h-5 w-5 text-blue-600" />
                        </div>
                        <h3 className="font-semibold text-gray-900">Personal Information</h3>
                      </div>
                      <div className="space-y-4">
                        <div>
                          <p className="text-sm text-gray-500 mb-1">Full Name</p>
                          <p className="font-medium text-gray-900">{user.first_name} {user.last_name}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-500 mb-1">Username</p>
                          <p className="font-medium text-gray-900">@{user.username}</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>

                  {/* Contact Info */}
                  <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-gray-50/50 hover:shadow-xl transition-all duration-300">
                    <CardContent className="p-6">
                      <div className="flex items-center gap-3 mb-4">
                        <div className="p-2 bg-green-100 rounded-lg">
                          <Mail className="h-5 w-5 text-green-600" />
                        </div>
                        <h3 className="font-semibold text-gray-900">Contact</h3>
                      </div>
                      <div className="space-y-4">
                        <div>
                          <p className="text-sm text-gray-500 mb-1">Email Address</p>
                          <p className="font-medium text-gray-900">{user.email}</p>
                        </div>
                        <div>
                          <p className="text-sm text-gray-500 mb-1">Member Since</p>
                          <p className="font-medium text-gray-900">{format(new Date(user.created_at), 'MMMM dd, yyyy')}</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </div>
              )}

              {/* Preferences */}
              <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-purple-50/30">
                <CardHeader className="pb-6">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-purple-100 rounded-lg">
                      <Settings className="h-5 w-5 text-purple-600" />
                    </div>
                    <div>
                      <CardTitle className="text-xl font-semibold text-gray-900">Preferences</CardTitle>
                      <CardDescription className="text-gray-600">
                        Customize your news experience
                      </CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-6">
                    <div className="flex items-center justify-between p-4 bg-white/60 rounded-xl border border-gray-100">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Bell className="h-4 w-4 text-blue-600" />
                          <Label className="text-base font-medium text-gray-900">Email Notifications</Label>
                        </div>
                        <p className="text-sm text-gray-600">
                          Receive email notifications for breaking news
                        </p>
                      </div>
                      <Switch
                        checked={user.preferences?.notification_enabled || false}
                        onCheckedChange={(checked) => {
                          // TODO: Update preference
                          console.log('Email notifications:', checked);
                        }}
                        className="data-[state=checked]:bg-blue-600"
                      />
                    </div>

                    <div className="flex items-center justify-between p-4 bg-white/60 rounded-xl border border-gray-100">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Mail className="h-4 w-4 text-green-600" />
                          <Label className="text-base font-medium text-gray-900">Daily Email Digest</Label>
                        </div>
                        <p className="text-sm text-gray-600">
                          Get a daily summary of top stories
                        </p>
                      </div>
                      <Switch
                        checked={user.preferences?.email_digest || false}
                        onCheckedChange={(checked) => {
                          // TODO: Update preference
                          console.log('Email digest:', checked);
                        }}
                        className="data-[state=checked]:bg-green-600"
                      />
                    </div>

                    <div className="flex items-center justify-between p-4 bg-white/60 rounded-xl border border-gray-100">
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Globe className="h-4 w-4 text-purple-600" />
                          <Label className="text-base font-medium text-gray-900">Theme Preference</Label>
                        </div>
                        <p className="text-sm text-gray-600">
                          Choose your preferred theme
                        </p>
                      </div>
                      <Badge variant="outline" className="bg-white/80">
                        {user.preferences?.theme || 'Light'}
                      </Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Sidebar */}
            <div className="space-y-8">
              {/* Activity Stats */}
              <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-blue-50/50 overflow-hidden">
                <CardHeader className="pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-blue-100 rounded-lg">
                      <Activity className="h-5 w-5 text-blue-600" />
                    </div>
                    <CardTitle className="text-lg font-semibold text-gray-900">Activity Overview</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="space-y-6">
                  <div className="space-y-4">
                    <div className="flex items-center justify-between p-4 bg-white/70 rounded-xl border border-blue-100">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-blue-500 rounded-lg">
                          <BookOpen className="h-4 w-4 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-gray-600">Articles Read</p>
                          <p className="text-2xl font-bold text-gray-900">0</p>
                        </div>
                      </div>
                      <div className="text-xs text-blue-600 font-medium bg-blue-50 px-2 py-1 rounded-full">
                        This Week
                      </div>
                    </div>
                    
                    <div className="flex items-center justify-between p-4 bg-white/70 rounded-xl border border-red-100">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-red-500 rounded-lg">
                          <Heart className="h-4 w-4 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-gray-600">Bookmarks</p>
                          <p className="text-2xl font-bold text-gray-900">{authState.bookmarks.length}</p>
                        </div>
                      </div>
                      <div className="text-xs text-red-600 font-medium bg-red-50 px-2 py-1 rounded-full">
                        Total
                      </div>
                    </div>

                    <div className="flex items-center justify-between p-4 bg-white/70 rounded-xl border border-green-100">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-green-500 rounded-lg">
                          <TrendingUp className="h-4 w-4 text-white" />
                        </div>
                        <div>
                          <p className="text-sm font-medium text-gray-600">Trending Clicks</p>
                          <p className="text-2xl font-bold text-gray-900">0</p>
                        </div>
                      </div>
                      <div className="text-xs text-green-600 font-medium bg-green-50 px-2 py-1 rounded-full">
                        Today
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Quick Actions */}
              <Card className="border-0 shadow-lg bg-gradient-to-br from-white to-gray-50">
                <CardHeader className="pb-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-gray-100 rounded-lg">
                      <Settings className="h-5 w-5 text-gray-600" />
                    </div>
                    <CardTitle className="text-lg font-semibold text-gray-900">Quick Actions</CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  <Button 
                    variant="outline" 
                    className="w-full justify-start bg-white/70 hover:bg-white hover:shadow-md border-gray-200 transition-all duration-200"
                  >
                    <Bell className="h-4 w-4 mr-3 text-blue-600" />
                    <span className="text-gray-700">Notification Settings</span>
                  </Button>
                  <Button 
                    variant="outline" 
                    className="w-full justify-start bg-white/70 hover:bg-white hover:shadow-md border-gray-200 transition-all duration-200"
                    onClick={() => navigate('/bookmarks')}
                  >
                    <Heart className="h-4 w-4 mr-3 text-red-600" />
                    <span className="text-gray-700">View Bookmarks ({authState.bookmarks.length})</span>
                  </Button>
                  <Button 
                    variant="outline" 
                    className="w-full justify-start bg-white/70 hover:bg-white hover:shadow-md border-gray-200 transition-all duration-200"
                  >
                    <Settings className="h-4 w-4 mr-3 text-gray-600" />
                    <span className="text-gray-700">Account Settings</span>
                  </Button>
                  <Button 
                    variant="outline" 
                    className="w-full justify-start bg-white/70 hover:bg-white hover:shadow-md border-gray-200 transition-all duration-200"
                  >
                    <Clock className="h-4 w-4 mr-3 text-purple-600" />
                    <span className="text-gray-700">Reading History</span>
                  </Button>
                </CardContent>
              </Card>

              {/* Account Status */}
              <Card className="border-0 shadow-lg bg-gradient-to-br from-green-50 to-emerald-50">
                <CardContent className="p-6">
                  <div className="text-center space-y-3">
                    <div className="inline-flex items-center justify-center w-12 h-12 bg-green-100 rounded-full">
                      <Shield className="h-6 w-6 text-green-600" />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-900">Account Verified</h3>
                      <p className="text-sm text-gray-600">Your account is secure and verified</p>
                    </div>
                    <div className="flex items-center justify-center gap-2 text-xs text-green-600 font-medium">
                      <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                      All systems operational
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile;
