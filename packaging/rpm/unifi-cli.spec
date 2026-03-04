Name:           unifi-cli
Version:        1.1.0
Release:        1%{?dist}
Summary:        UniFi Controller CLI tool for local network management
License:        MIT
URL:            https://github.com/dl-alexandre/Local-UniFi-CLI
Source0:        %{name}-%{version}.tar.gz

%description
Command-line interface for managing UniFi network controllers.
Supports local controller management, device monitoring, and network configuration.

%prep
%setup -q

%build
go build -o unifi ./cmd/unifi

%install
mkdir -p %{buildroot}/%{_bindir}
cp unifi %{buildroot}/%{_bindir}/

%files
%{_bindir}/unifi
%doc README.md
%license LICENSE

%changelog
* Mon Jan 01 2024 Dalton Alexandre <dalexandre@milcgroup.info> - 1.1.0-1
- Initial RPM release
